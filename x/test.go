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
	Results   Results
	Result    TestCaseResult
}

// TestCaseResult represents an expectation for TestCase.Results.
type TestCaseResult interface {
	start(t *testing.T, m *stackvm.Mach) finisher
	startTraced(t *testing.T, m *stackvm.Mach) finisher
}

type finisher interface {
	finish(m *stackvm.Mach)
}

type testCaseRun struct {
	*testing.T
	TestCase
	res []Result
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
	if t.Results != nil {
		m.SetHandler(t.queueSize(), stackvm.HandlerFunc(t.takeResult))
	}
	t.checkError(m.Run())
	if t.Results == nil {
		assert.Nil(t, t.res, "unexpected results")
	} else {
		assert.Equal(t, t.Results, t.res, "expected results")
	}
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
	fin := t.Result.startTraced(t.T, m)
	if t.Results != nil {
		m.SetHandler(t.queueSize(), stackvm.HandlerFunc(t.checkEachResult))
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

// Result represents an expected or actual result within a TestCase.
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

func (r Result) start(t *testing.T, m *stackvm.Mach) finisher       { return runResult{t, r} }
func (r Result) startTraced(t *testing.T, m *stackvm.Mach) finisher { return r.start(t, m) }

type runResult struct {
	*testing.T
	Result
}

func (rr runResult) finish(m *stackvm.Mach) {
	actual, err := rr.Result.take(m)
	assert.NoError(rr, err, "unexpected error taking final result")
	assert.Equal(rr, rr.Result, actual, "expected result")
}

// Results represents multiple expected results.
type Results []Result

func (t *testCaseRun) checkEachResult(m *stackvm.Mach) error {
	i, expected, actual, err := t._takeResult(m)
	if err != nil {
		return err
	}
	if i >= len(t.Results) {
		assert.Fail(t, "unexpected result", "unexpected result[%d]: %+v", i, actual)
	} else {
		assert.Equal(t, expected, actual, "expected result[%d]", i)
		// TODO if { note(m, "^^^", "expected result", "result[%d] == %+v", i, actual) }
	}
	return nil
}

func (t *testCaseRun) _takeResult(m *stackvm.Mach) (i int, expected, actual Result, err error) {
	i = len(t.res)
	if i < len(t.Results) {
		expected = t.Results[i]
	}
	actual, err = expected.take(m)
	t.res = append(t.res, actual)
	return
}

func (t *testCaseRun) takeResult(m *stackvm.Mach) error {
	_, _, _, err := t._takeResult(m)
	return err
}

type finishers []finisher

func (fs finishers) finish(m *stackvm.Mach) {
	for i := range fs {
		fs[i].finish(m)
	}
}
