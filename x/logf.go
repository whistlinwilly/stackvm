package xstackvm

import (
	"fmt"

	"github.com/jcorbin/stackvm"
)

const noteWidth = 15

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
	lf.note(m, "===", "Begin", "stacks=[0x%04x:0x%04x]", m.PBP(), m.CBP())
}

// End logs end of machine run (before any handling).
func (lf *LogfTracer) End(m *stackvm.Mach) {
	lf.note(m, "===", "End")
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
	mid := lf.machID(m)
	lf.machID(n)
	lf.note(n, "+++", fmt.Sprintf("%v copy", mid))
}

// Handle logs any handling error.
func (lf *LogfTracer) Handle(m *stackvm.Mach, err error) {
	if err != nil {
		lf.note(m, "!!!", err)
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
	format := "%v %s % *v @0x%04x"
	parts := []interface{}{lf.ids[m], mark, noteWidth, note, m.IP()}
	if len(args) > 0 {
		if s, ok := args[0].(string); ok {
			format += " " + s
		}
		parts = append(parts, args[1:]...)
	}
	lf.f(format, parts...)
}
