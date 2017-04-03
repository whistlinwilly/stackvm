package xstackvm

import (
	"fmt"

	"github.com/jcorbin/stackvm"
	"github.com/jcorbin/stackvm/x/action"
	"github.com/jcorbin/stackvm/x/dumper"
	"github.com/jcorbin/stackvm/x/tracer"
)

// NewLogfTracer creates a tracer that logs a trace of the machines' exepciton.
func NewLogfTracer(f func(string, ...interface{})) stackvm.Tracer {
	return tracer.Multi(
		tracer.NewIDTracer(),
		tracer.NewCountTracer(),
		tracer.NewLogTracer(f),
		&logfTracer{
			dmw: action.Never,
		},
	)
}

type logfTracer struct {
	dmw action.Predicate
}

func (lf *logfTracer) Context(m *stackvm.Mach, key string) (interface{}, bool) { return nil, false }

func (lf *logfTracer) Begin(m *stackvm.Mach) {
	if lf.dmw.Test(action.TraceBegin, 0, stackvm.Op{}) {
		dumpMem(m)
	}
}

func (lf *logfTracer) End(m *stackvm.Mach) {
	if lf.dmw.Test(action.TraceEnd, 0, stackvm.Op{}) {
		dumpMem(m)
	}
}

func (lf *logfTracer) Before(m *stackvm.Mach, ip uint32, op stackvm.Op) {
	if lf.dmw.Test(action.TraceBefore, ip, op) {
		dumpMem(m)
	}
}

func (lf *logfTracer) After(m *stackvm.Mach, ip uint32, op stackvm.Op) {
	if lf.dmw.Test(action.TraceAfter, ip, op) {
		dumpMem(m)
	}
}

func (lf *logfTracer) Queue(m, n *stackvm.Mach) {
	if lf.dmw.Test(action.TraceQueue, 0, stackvm.Op{}) {
		dumpMem(m)
	}
}

func (lf *logfTracer) Handle(m *stackvm.Mach, err error) {
	if lf.dmw.Test(action.TraceHandle, 0, stackvm.Op{}) {
		dumpMem(m)
	}
}

func defaultLogf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

func dumpMem(m *stackvm.Mach) {
	logf := defaultLogf
	if v, def := m.Tracer().Context(m, "logf"); def {
		if f, ok := v.(func(string, ...interface{})); ok {
			logf = f
		}
	}
	dumper.Dump(m, logf)
}
