package xstackvm

import (
	"flag"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcorbin/stackvm"
	"github.com/jcorbin/stackvm/internal/errors"
	"github.com/jcorbin/stackvm/x/action"
	"github.com/jcorbin/stackvm/x/dumper"
	"github.com/jcorbin/stackvm/x/tracer"
)

var (
	traceFlag   bool
	dumpMemFlag action.PredicateFlag
)

func init() {
	flag.BoolVar(&traceFlag, "stackvm.test.trace", false,
		"run any stackvm tests with tracing on, even if they pass")
	flag.Var(&dumpMemFlag, "stackvm.test.dumpmem",
		"dump memory when the given predicates are true (FIXME predicates?!?)")
}

// TestCases is list of test cases for stackvm.
type TestCases []TestCase

// TestCase is a test case for a stackvm.
type TestCase struct {
	Logf      func(format string, args ...interface{})
	Name      string
	Prog      []byte
	Err       string
	QueueSize int
	Handler   func(*stackvm.Mach) ([]byte, error)
	Result    TestCaseResult
}

// TestCaseResult represents an expectation for TestCase.Results.  Both of the
// Result and Results types implement this interface, and can be used directly
// to express simple expectations.
type TestCaseResult interface {
	start(t *testing.T, m *stackvm.Mach) finisher
}

type finisher interface {
	finish(m *stackvm.Mach)
}

type testCaseRun struct {
	*testing.T
	TestCase
}

// Run runs each test case in a sub-test.
func (tcs TestCases) Run(t *testing.T) {
	for _, tc := range tcs {
		t.Run(tc.Name, tc.Run)
	}
}

// Trace traces each test case in a sub-test.
func (tcs TestCases) Trace(t *testing.T) {
	for _, tc := range tcs {
		t.Run(tc.Name, tc.Trace)
	}
}

// TraceTo traces each test case in a sub-test.
func (tcs TestCases) TraceTo(t *testing.T, w io.Writer) {
	for _, tc := range tcs {
		t.Run(tc.Name, tc.LogTo(w).Trace)
	}
}

type ioLogger struct {
	err error
	w   io.Writer
}

func (iol *ioLogger) logf(format string, args ...interface{}) {
	if iol.err != nil {
		return
	}
	if _, err := fmt.Fprintf(iol.w, format+"\n", args...); err != nil {
		iol.err = err
	}
}

// LogTo returns a copy of the test case with Logf
// changed to print to the given io.Writer.
func (tc TestCase) LogTo(w io.Writer) TestCase {
	iol := ioLogger{w: w}
	tc.Logf = iol.logf
	return tc
}

// Run runs the test case; it either succeeds quietly, or fails with a trace
// log.
func (tc TestCase) Run(t *testing.T) {
	run := testCaseRun{
		T:        t,
		TestCase: tc,
	}
	if traceFlag || run.canaryFailed() {
		run.trace()
	}
}

// Trace runs the test case with trace logging on.
func (tc TestCase) Trace(t *testing.T) {
	run := testCaseRun{
		T:        t,
		TestCase: tc,
	}
	run.trace()
}

func (t testCaseRun) contextLog(m *stackvm.Mach) func(string, ...interface{}) {
	logf := t.Logf
	if v, def := m.Tracer().Context(m, "logf"); def {
		if f, ok := v.(func(string, ...interface{})); ok {
			logf = f
		}
	}
	return logf
}

func (t testCaseRun) queueSize() int {
	if t.QueueSize <= 0 {
		return 10
	}
	return t.QueueSize
}

func (t testCaseRun) canaryFailed() bool {
	if t.Logf == nil {
		t.Logf = t.T.Logf
	}
	t.T = &testing.T{}
	m := t.build()
	fin := t.Result.start(t.T, m)
	if h, ok := fin.(stackvm.Handler); ok {
		m.SetHandler(t.queueSize(), h)
	}
	t.checkError(m.Run())
	fin.finish(m)
	return t.Failed()
}

func (t testCaseRun) trace() {
	if t.Logf == nil {
		t.Logf = t.T.Logf
	}
	trc := tracer.Multi(
		tracer.NewIDTracer(),
		tracer.NewCountTracer(),
		tracer.NewLogTracer(t.Logf),
		tracer.Filtered(
			tracer.FuncTracer(func(m *stackvm.Mach) {
				_ = dumper.Dump(m, t.contextLog(m))
			}),
			dumpMemFlag.Build(),
		),
	)

	m := t.build()
	fin := t.Result.start(t.T, m)
	if h, ok := fin.(stackvm.Handler); ok {
		m.SetHandler(t.queueSize(), h)
	}
	t.checkError(m.Trace(trc))
	fin.finish(m)
}

func (t testCaseRun) logLines(s string) {
	for _, line := range strings.Split(s, "\n") {
		t.Logf(line)
	}
}

func (t testCaseRun) build() *stackvm.Mach {
	m, err := stackvm.New(t.Prog)
	require.NoError(t, err, "unexpected machine compile error")
	return m
}

func (t testCaseRun) checkError(err error) {
	if t.Err == "" {
		assert.NoError(t, err, "unexpected run error")
	} else {
		assert.EqualError(t, errors.Cause(err), t.Err, "unexpected run error")
	}
}

// Result represents an expected or actual result within a TestCase. It can be
// used as a TestCaseResult when only a single final result is expected.
type Result struct {
	Err    string
	Values [][]uint32
}

func (r Result) take(m *stackvm.Mach) (res Result, err error) {
	if merr := m.Err(); merr != nil {
		res.Err = errors.Cause(merr).Error()
	} else {
		res.Values, err = m.Values()
	}
	return
}

func (r Result) start(t *testing.T, m *stackvm.Mach) finisher { return runResult{t, r} }

type runResult struct {
	*testing.T
	Result
}

func (rr runResult) finish(m *stackvm.Mach) {
	actual, err := rr.Result.take(m)
	assert.NoError(rr, err, "unexpected error taking final result")
	assert.Equal(rr, rr.Result, actual, "expected result")
}

// Results represents multiple expected results. It can be used as a
// TestCaseResult to require an exact sequence of results, including failed
// non-zero halt states.
type Results []Result

func (rs Results) start(t *testing.T, m *stackvm.Mach) finisher {
	return &runResults{t, rs, 0}
}

type runResults struct {
	*testing.T
	expected Results
	i        int
}

func (rrs *runResults) Handle(m *stackvm.Mach) error {
	var expected Result
	i := rrs.i
	if i < len(rrs.expected) {
		expected = rrs.expected[i]
	}
	actual, err := expected.take(m)
	rrs.i++
	if err != nil {
		return err
	}
	if i >= len(rrs.expected) {
		assert.Fail(rrs, "unexpected result", "unexpected result[%d]: %+v", i, actual)
	} else {
		assert.Equal(rrs, expected, actual, "expected result[%d]", i)
		// TODO if { note(m, "^^^", "expected result", "result[%d] == %+v", i, actual) }
	}
	return nil
}

func (rrs *runResults) finish(m *stackvm.Mach) {
}
