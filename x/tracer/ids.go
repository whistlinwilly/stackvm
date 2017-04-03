package tracer

import (
	"fmt"

	"github.com/jcorbin/stackvm"
)

// MachID represents a 2-part machine id that tracks parentage. Each component
// is a number drawn from a monotonic counter. These numbers are chosen when
// the machine Begin()s execution, or when it is Queue()ed by its parent.
//
// The two components ids are called "parent" and "self": "parent" is set to 0
// when Begin() assigns an id, and inherited from the parent's "self" when
// assigned by Queue(); "self" is simply assigned to the next available id in
// both Queue() and Begin().
type MachID [2]int

func (mid MachID) String() string {
	return fmt.Sprintf("(%d:%d)", mid[0], mid[1])
}

// NewIDTracer creates a tracer that assigns MachIDs to machines.
func NewIDTracer() stackvm.Tracer {
	return &idTracer{
		ids: make(map[*stackvm.Mach]MachID),
	}
}

type idTracer struct {
	nextID int
	ids    map[*stackvm.Mach]MachID
}

func (it *idTracer) Context(m *stackvm.Mach, key string) (interface{}, bool) {
	if key != "id" {
		return nil, false
	}
	if id, def := it.ids[m]; def {
		return id, true
	}
	return nil, true
}

func (it *idTracer) Begin(m *stackvm.Mach) {
	if _, def := it.ids[m]; !def {
		it.nextID++
		it.ids[m] = MachID{0, it.nextID}
	}
}

func (it *idTracer) Before(m *stackvm.Mach, ip uint32, op stackvm.Op) {
}

func (it *idTracer) After(m *stackvm.Mach, ip uint32, op stackvm.Op) {
}

func (it *idTracer) Queue(m, n *stackvm.Mach) {
	delete(it.ids, n)
	mid, def := it.ids[m]
	if !def {
		it.nextID++
		mid = MachID{0, it.nextID}
		it.ids[m] = mid
	}
	it.nextID++
	it.ids[n] = MachID{mid[1], it.nextID}
}

func (it *idTracer) End(m *stackvm.Mach) {}

func (it *idTracer) Handle(m *stackvm.Mach, err error) {
	delete(it.ids, m)
}
