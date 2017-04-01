package action

import "github.com/jcorbin/stackvm"

// Predicate matches machine trace state.
type Predicate interface {
	Test(act TraceAction, ip uint32, op stackvm.Op) bool
}

// TODO: flatten same kind under Any/All, elide nils, and fold fixeds; maybe
// need to copy slice?

// Any returns a predicate that works as the logical Or of all the given
// predicates.
func Any(ps ...Predicate) Predicate {
	switch len(ps) {
	case 0:
		return nil
	case 1:
		return ps[0]
	default:
		return anyPredicate(ps)
	}
}

// All returns a predicate that works as the logical And of all the given
// predicates.
func All(ps ...Predicate) Predicate {
	switch len(ps) {
	case 0:
		return nil
	case 1:
		return ps[0]
	default:
		return allPredicate(ps)
	}
}

var (
	// Never is an always false predicate.
	Never = fixedPredicate(false)

	// Always is an always true predicate.
	Always = fixedPredicate(true)
)

type anyPredicate []Predicate
type allPredicate []Predicate
type fixedPredicate bool

func (b fixedPredicate) Test(_ TraceAction, _ uint32, _ stackvm.Op) bool { return bool(b) }

// TestFunc is a convenience for implementing Predicate directly with
// a function.
type TestFunc func(act TraceAction, ip uint32, op stackvm.Op) bool

// Test calls the wrapped function.
func (f TestFunc) Test(act TraceAction, ip uint32, op stackvm.Op) bool { return f(act, ip, op) }

func (any anyPredicate) Test(act TraceAction, ip uint32, op stackvm.Op) bool {
	for _, p := range any {
		if p.Test(act, ip, op) {
			return true
		}
	}
	return false
}

func (all allPredicate) Test(act TraceAction, ip uint32, op stackvm.Op) bool {
	for _, p := range all {
		if !p.Test(act, ip, op) {
			return false
		}
	}
	return true
}
