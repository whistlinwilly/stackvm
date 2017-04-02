package xstackvm

import (
	"fmt"

	"github.com/jcorbin/stackvm"
	"github.com/jcorbin/stackvm/x/action"
	"github.com/jcorbin/stackvm/x/dumper"
)

// NewLogfTracer creates a tracer that logs a trace of the machines' exepciton.
func NewLogfTracer(f func(string, ...interface{})) stackvm.Tracer {
	return &logfTracer{
		ids:   make(map[*stackvm.Mach]machID),
		count: make(map[*stackvm.Mach]int),
		f:     f,
		dmw:   action.Never,
	}
}

const noteWidth = 15

type logfTracer struct {
	nextID int
	ids    map[*stackvm.Mach]machID
	count  map[*stackvm.Mach]int
	f      func(string, ...interface{})
	dmw    action.Predicate
}

type machID [2]int

func (mi machID) String() string { return fmt.Sprintf("(%d:%d)", mi[0], mi[1]) }

func (lf *logfTracer) Context(m *stackvm.Mach, key string) (interface{}, bool) { return nil, false }

func (lf *logfTracer) Begin(m *stackvm.Mach) {
	lf.machID(m)
	lf.note(m, "===", "Begin", "stacks=[0x%04x:0x%04x]", m.PBP(), m.CBP())
	if lf.dmw.Test(action.TraceBegin, 0, stackvm.Op{}) {
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

func (lf *logfTracer) End(m *stackvm.Mach) {
	if err := m.Err(); err != nil {
		lf.note(m, "===", "End", "err=%v", cause(err))
	} else if vs, err := m.Values(); err != nil {
		lf.note(m, "===", "End", "values_err=%v", err)
	} else {
		lf.note(m, "===", "End", "values=%v", vs)
	}
	if lf.dmw.Test(action.TraceEnd, 0, stackvm.Op{}) {
		lf.dumpMem(m, "...")
	}
}

func (lf *logfTracer) Before(m *stackvm.Mach, ip uint32, op stackvm.Op) {
	lf.count[m]++
	lf.noteStack(m, ">>>", op)
	if lf.dmw.Test(action.TraceBefore, ip, op) {
		lf.dumpMem(m, "...")
	}
}

func (lf *logfTracer) After(m *stackvm.Mach, ip uint32, op stackvm.Op) {
	lf.noteStack(m, "...", "")
	if lf.dmw.Test(action.TraceAfter, ip, op) {
		lf.dumpMem(m, "...")
	}
}

func (lf *logfTracer) Queue(m, n *stackvm.Mach) {
	delete(lf.ids, n)
	mid := lf.machID(m)
	lf.count[n] = lf.count[m]
	lf.nextID++
	lf.ids[n] = machID{mid[1], lf.nextID}
	lf.note(n, "+++", fmt.Sprintf("%v copy", mid))
	if lf.dmw.Test(action.TraceQueue, 0, stackvm.Op{}) {
		lf.dumpMem(m, "...")
	}
}

func (lf *logfTracer) Handle(m *stackvm.Mach, err error) {
	if err != nil {
		lf.note(m, "!!!", err)
	}

	delete(lf.ids, m)
	delete(lf.count, m)
}

func (lf *logfTracer) machID(m *stackvm.Mach) machID {
	id, def := lf.ids[m]
	if !def {
		lf.nextID++
		id = machID{0, lf.nextID}
		lf.ids[m] = id
	}
	return id
}

func (lf *logfTracer) noteStack(m *stackvm.Mach, mark string, note interface{}) {
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

func (lf *logfTracer) note(m *stackvm.Mach, mark string, note interface{}, args ...interface{}) {
	format := "%v #% 4d %s % *v @0x%04x"
	parts := []interface{}{lf.ids[m], lf.count[m], mark, noteWidth, note, m.IP()}
	if len(args) > 0 {
		if s, ok := args[0].(string); ok {
			format += " " + s
			args = args[1:]
		}
		parts = append(parts, args...)
	}
	lf.f(format, parts...)
}

func (lf *logfTracer) dumpMem(m *stackvm.Mach, mark string) {
	pfx := []interface{}{lf.ids[m], mark}
	dumper.Dump(m, func(format string, args ...interface{}) {
		format = "%v       %s " + format
		args = append(pfx[:2:2], args...)
		lf.f(format, args...)
	})
}
