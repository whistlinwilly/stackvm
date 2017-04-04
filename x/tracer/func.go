package tracer

import "github.com/jcorbin/stackvm"

// FuncTracer creates a tracer that just calls a given function, passing it the
// machine being traced.
func FuncTracer(f func(*stackvm.Mach)) stackvm.Tracer {
	return funcTracer(f)
}

type funcTracer func(*stackvm.Mach)

func (ft funcTracer) Context(m *stackvm.Mach, key string) (interface{}, bool) { return nil, false }
func (ft funcTracer) Begin(m *stackvm.Mach)                                   { ft(m) }
func (ft funcTracer) End(m *stackvm.Mach)                                     { ft(m) }
func (ft funcTracer) Queue(m, n *stackvm.Mach)                                { ft(m) }
func (ft funcTracer) Handle(m *stackvm.Mach, err error)                       { ft(m) }
func (ft funcTracer) Before(m *stackvm.Mach, ip uint32, op stackvm.Op)        { ft(m) }
func (ft funcTracer) After(m *stackvm.Mach, ip uint32, op stackvm.Op)         { ft(m) }
