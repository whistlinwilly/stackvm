package xstackvm

import "github.com/jcorbin/stackvm"

// LogfTracer implements a simple stackvm.Tracer that prints to a
// formatted log function.
type LogfTracer func(string, ...interface{})

// Begin logs start of machine run.
func (logf LogfTracer) Begin(m *stackvm.Mach) {
	logf("BEGIN %v", m)
}

// End logs end of machine run (before any handling).
func (logf LogfTracer) End(m *stackvm.Mach) {
	logf("END %v", m)
}

// Before logs an operation about to be executed.
func (logf LogfTracer) Before(m *stackvm.Mach, _ uint32, op stackvm.Op) {
	logf(">>> % 10v in %v", op, m)
}

// After logs the result of executing an operation.
func (logf LogfTracer) After(m *stackvm.Mach, ip uint32, op stackvm.Op) {
	logf("... % 10v in %v", op, m)
}

// Queue logs a copy of a machine being ran.
func (logf LogfTracer) Queue(m, n *stackvm.Mach) {
	logf("+++ %v", n)
}

// Handle logs any handling error.
func (logf LogfTracer) Handle(m *stackvm.Mach, err error) {
	if err != nil {
		logf("ERR %v %v", m, err)
	}
}
