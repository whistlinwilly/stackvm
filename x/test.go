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
	Results   []Result
	Result    Result
}

// Result represents an expected or actual result within a TestCase.
type Result struct {
	Err    string
	Values [][]uint32
}

// ResultMem represents an expected or actual memory range in a Result.
type ResultMem struct {
	Addr uint32
	Data []byte
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

func (t testCaseRun) note(m *stackvm.Mach, format string, args ...interface{}) {
	// TODO: better looking if we can use log tracer's note function
	t.contextLog(m)(
		"@0x%04x "+format,
		append([]interface{}{m.IP()}, args...)...)
}

func (t testCaseRun) canaryFailed() bool {
	if t.Logf == nil {
		t.Logf = t.T.Logf
	}
	t.T = &testing.T{}
	m := t.build(t.takeResult)
	t.checkError(m.Run())
	t.checkResults(m, true)
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

	m := t.build(t.checkEachResult)
	t.checkError(m.Trace(trc))
	t.checkResults(m, false)
}

func (t testCaseRun) logLines(s string) {
	for _, line := range strings.Split(s, "\n") {
		t.Logf(line)
	}
}

func (t testCaseRun) build(handle func(*stackvm.Mach) error) *stackvm.Mach {
	m, err := stackvm.New(t.Prog)
	require.NoError(t, err, "unexpected machine compile error")
	if t.Results != nil {
		qs := t.QueueSize
		if qs <= 0 {
			qs = 10
		}
		m.SetHandler(qs, handle)
	}
	return m
}

func (t testCaseRun) checkError(err error) {
	if t.Err == "" {
		assert.NoError(t, err, "unexpected run error")
	} else {
		assert.EqualError(t, errors.Cause(err), t.Err, "unexpected run error")
	}
}

func (t testCaseRun) checkResults(m *stackvm.Mach, resultsTaken bool) {
	if t.Results == nil {
		assert.Nil(t, t.res, "unexpected results")
	} else if resultsTaken {
		assert.Equal(t, t.Results, t.res, "expected results")
	}
	t.checkFinalResult(m)
}

func (t testCaseRun) checkFinalResult(m *stackvm.Mach) {
	actual, err := t.Result.take(m)
	assert.NoError(t, err, "unexpected error taking final result")
	assert.Equal(t, t.Result, actual, "expected result")
}

func (t *testCaseRun) checkEachResult(m *stackvm.Mach) error {
	i, expected, actual, err := t._takeResult(m)
	if err != nil {
		return err
	}
	if i >= len(t.Results) {
		assert.Fail(t, "unexpected result", "unexpected result[%d]: %+v", i, actual)
	} else if assert.Equal(t, expected, actual, "expected result[%d]", i) {
		t.note(m, "result[%d] == %+v", i, actual)
		// TODO: real note "^^^", "expected result",
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

func (r Result) take(m *stackvm.Mach) (res Result, err error) {
	if merr := m.Err(); merr != nil {
		res.Err = errors.Cause(merr).Error()
	} else {
		res.Values, err = m.Values()
	}
	return
}
