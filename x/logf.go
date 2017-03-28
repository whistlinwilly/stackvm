package xstackvm

import (
	"fmt"

	"github.com/jcorbin/stackvm"
)

// TraceAction represents one of the tracer methods.
type TraceAction int

const (
	// TraceBegin corresponds to Tracer.Begin.
	TraceBegin = TraceAction(iota + 1)
	// TraceEnd corresponds to Tracer.End.
	TraceEnd
	// TraceQueue corresponds to Tracer.Queue.
	TraceQueue
	// TraceBefore corresponds to Tracer.Before.
	TraceBefore
	// TraceAfter corresponds to Tracer.After.
	TraceAfter
)

const noteWidth = 15

// LogfTracer implements a stackvm.Tracer that prints to a formatted log
// function.
type LogfTracer struct {
	nextID int
	ids    map[*stackvm.Mach]machID
	count  map[*stackvm.Mach]int
	f      func(string, ...interface{})
	dmw    func(act TraceAction, ip uint32, op stackvm.Op) bool
}

type machID [2]int

func neverDumpMem(_ TraceAction, _ uint32, _ stackvm.Op) bool { return false }

func (mi machID) String() string { return fmt.Sprintf("(%d:%d)", mi[0], mi[1]) }

// NewLogfTracer creates a new tracer around a log formatting function.
func NewLogfTracer(f func(string, ...interface{})) *LogfTracer {
	return &LogfTracer{
		ids:   make(map[*stackvm.Mach]machID),
		count: make(map[*stackvm.Mach]int),
		f:     f,
		dmw:   neverDumpMem,
	}
}

// DumpMemWhen sets a predicate function that gets called in after: if it
// returns true, then the machine's memory is dumped.
func (lf *LogfTracer) DumpMemWhen(f func(act TraceAction, ip uint32, op stackvm.Op) bool) {
	lf.dmw = f
}

// Begin logs start of machine run.
func (lf *LogfTracer) Begin(m *stackvm.Mach) {
	lf.machID(m)
	lf.note(m, "===", "Begin", "stacks=[0x%04x:0x%04x]", m.PBP(), m.CBP())
	if lf.dmw(TraceBegin, 0, stackvm.Op{}) {
		lf.dumpMem(m, "...")
	}
}

type causer interface {
	Cause() error
}

func cause(err error) error {
	for {
		if c, ok := err.(causer); ok {
			err = c.Cause()
			continue
		}
		return err
	}
}

// End logs end of machine run (before any handling).
func (lf *LogfTracer) End(m *stackvm.Mach) {
	err := m.Err()
	lf.note(m, "===", "End", "err=%v", cause(err))
	delete(lf.ids, m)
	delete(lf.count, m)
	if lf.dmw(TraceEnd, 0, stackvm.Op{}) {
		lf.dumpMem(m, "...")
	}
}

// Before logs an operation about to be executed.
func (lf *LogfTracer) Before(m *stackvm.Mach, ip uint32, op stackvm.Op) {
	lf.count[m]++
	lf.noteStack(m, ">>>", op)
	if lf.dmw(TraceBefore, ip, op) {
		lf.dumpMem(m, "...")
	}
}

// After logs the result of executing an operation.
func (lf *LogfTracer) After(m *stackvm.Mach, ip uint32, op stackvm.Op) {
	lf.noteStack(m, "...", "")
	if lf.dmw(TraceAfter, ip, op) {
		lf.dumpMem(m, "...")
	}
}

// Queue logs a copy of a machine being ran.
func (lf *LogfTracer) Queue(m, n *stackvm.Mach) {
	delete(lf.ids, n)
	mid := lf.machID(m)
	lf.count[n] = lf.count[m]
	lf.machID(n)
	lf.note(n, "+++", fmt.Sprintf("%v copy", mid))
	if lf.dmw(TraceQueue, 0, stackvm.Op{}) {
		lf.dumpMem(m, "...")
	}
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
	format := "%v #% 4d %s % *v @0x%04x"
	parts := []interface{}{lf.ids[m], lf.count[m], mark, noteWidth, note, m.IP()}
	if len(args) > 0 {
		if s, ok := args[0].(string); ok {
			format += " " + s
		}
		parts = append(parts, args[1:]...)
	}
	lf.f(format, parts...)
}

func (lf *LogfTracer) dumpMem(m *stackvm.Mach, mark string) {
	pfx := []interface{}{lf.ids[m], mark}
	Dump(m, func(format string, args ...interface{}) {
		format = "%v       %s " + format
		args = append(pfx[:2:2], args...)
		lf.f(format, args...)
	})
}
