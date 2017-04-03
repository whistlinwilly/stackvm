package tracer

import "github.com/jcorbin/stackvm"

// NewCountTracer creates a tracer that counts machine operations. Counts are
// tracked by pointer, and deleted after the ended machine has been handled.
func NewCountTracer() stackvm.Tracer {
	return make(countTracer)
}

type countTracer map[*stackvm.Mach]int

func (ct countTracer) Context(m *stackvm.Mach, key string) (interface{}, bool) {
	if key != "count" {
		return nil, false
	}
	if c, def := ct[m]; def {
		return c, true
	}
	return nil, true
}

func (ct countTracer) Begin(m *stackvm.Mach) {}
func (ct countTracer) End(m *stackvm.Mach)   {}

func (ct countTracer) Before(m *stackvm.Mach, ip uint32, op stackvm.Op) {
	ct[m]++
}

func (ct countTracer) After(m *stackvm.Mach, ip uint32, op stackvm.Op) {
	// TODO: why not inc here?
}

func (ct countTracer) Queue(m, n *stackvm.Mach) {
	ct[n] = ct[m]
}

func (ct countTracer) Handle(m *stackvm.Mach, err error) {
	delete(ct, m)
}
