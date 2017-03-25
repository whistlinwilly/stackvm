package xstackvm

import "github.com/jcorbin/stackvm"

const opWidth = 15

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
func (logf LogfTracer) Before(m *stackvm.Mach, ip uint32, op stackvm.Op) {
	ps, cs, err := m.Stacks()
	if err != nil {
		logf(">>> % *v @0x%04x 0x%04x:0x%04x 0x%04x:0x%04x ERROR %v", opWidth, op, ip, m.PBP(), m.PSP(), m.CSP(), m.CBP(), err)
	} else {
		logf(">>> % *v @0x%04x %v :0x%04x 0x%04x: %v", opWidth, op, ip, ps, m.PSP(), m.CSP(), cs)
	}
}

// After logs the result of executing an operation.
func (logf LogfTracer) After(m *stackvm.Mach, ip uint32, op stackvm.Op) {
	ps, cs, err := m.Stacks()
	if err != nil {
		logf("... % *v @0x%04x 0x%04x:0x%04x 0x%04x:0x%04x ERROR %v", opWidth, "", ip, m.PBP(), m.PSP(), m.CSP(), m.CBP(), err)
	} else {
		logf("... % *v @0x%04x %v :0x%04x 0x%04x: %v", opWidth, "", ip, ps, m.PSP(), m.CSP(), cs)
	}
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
