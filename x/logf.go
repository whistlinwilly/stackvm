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
		tracer.Filtered(
			tracer.FuncTracer(dumpMem),
			action.Never,
		),
	)
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
