package xstackvm

import (
	"fmt"

	"github.com/jcorbin/stackvm"
)

const opWidth = 15

// LogfTracer implements a stackvm.Tracer that prints to a formatted log
// function.
type LogfTracer struct {
	nextID int
	ids    map[*stackvm.Mach]machID
	f      func(string, ...interface{})
}

type machID [2]int

func (mi machID) String() string { return fmt.Sprintf("(%d:%d)", mi[0], mi[1]) }

// NewLogfTracer creates a new tracer around a log formatting function.
func NewLogfTracer(f func(string, ...interface{})) *LogfTracer {
	return &LogfTracer{
		ids: make(map[*stackvm.Mach]machID),
		f:   f,
	}
}

// Begin logs start of machine run.
func (lf *LogfTracer) Begin(m *stackvm.Mach) {
	lf.machID(m)
	lf.logf(m, "BEGIN @0x%04x", m.IP())
}

// End logs end of machine run (before any handling).
func (lf *LogfTracer) End(m *stackvm.Mach) {
	lf.logf(m, "END")
	delete(lf.ids, m)
}

// Before logs an operation about to be executed.
func (lf *LogfTracer) Before(m *stackvm.Mach, ip uint32, op stackvm.Op) {
	lf.noteStack(m, ">>>", op)
}

// After logs the result of executing an operation.
func (lf *LogfTracer) After(m *stackvm.Mach, ip uint32, op stackvm.Op) {
	lf.noteStack(m, "...", "")
}

// Queue logs a copy of a machine being ran.
func (lf *LogfTracer) Queue(m, n *stackvm.Mach) {
	delete(lf.ids, n)
	lf.machID(m)
	nid := lf.machID(n)
	lf.logf(m, "+++ %v %v", nid, n)
}

// Handle logs any handling error.
func (lf *LogfTracer) Handle(m *stackvm.Mach, err error) {
	if err != nil {
		lf.logf(m, "ERR %v", err)
	}
}

func (lf *LogfTracer) machID(m *stackvm.Mach) machID {
	id, def := lf.ids[m]
	if !def {
		lf.nextID++
		id = machID{0, lf.nextID}
		lf.ids[m] = id
	}
	return id
}

func (lf *LogfTracer) noteStack(m *stackvm.Mach, mark string, note interface{}) {
	ps, cs, err := m.Stacks()
	if err != nil {
		lf.note(m, mark, note,
			"0x%04x:0x%04x 0x%04x:0x%04x ERROR %v",
			m.PBP(), m.PSP(), m.CSP(), m.CBP(), err)
	} else {
		lf.note(m, mark, note,
			"%v :0x%04x 0x%04x: %v",
			ps, m.PSP(), m.CSP(), cs)
	}
}

func (lf *LogfTracer) note(m *stackvm.Mach, mark string, note interface{}, args ...interface{}) {
	format := "%s % *v @0x%04x"
	parts := []interface{}{mark, opWidth, note, m.IP()}
	if len(args) > 0 {
		if s, ok := args[0].(string); ok {
			format += " " + s
		}
		parts = append(parts, args[1:]...)
	}
	lf.logf(m, format, parts...)
}

func (lf *LogfTracer) logf(m *stackvm.Mach, format string, args ...interface{}) {
	format = "%v " + format
	id := lf.ids[m]
	lf.f(format, append([]interface{}{id}, args...)...)
}
