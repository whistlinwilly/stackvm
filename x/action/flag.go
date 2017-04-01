package action

import "strings"

// PredicateFlag collects predicates specified by a command line flag.
// Predicates strings may be given over multiple flag instances, or as
// space-separated fields within one flag value. All such collected predicates
// are combined into an Any() Predicate; i.e. implicit "or" semantics.
type PredicateFlag struct {
	ps []Predicate
}

func (pf *PredicateFlag) String() string {
	switch len(pf.ps) {
	case 0:
		return ""
	case 1:
		return PredicateString(pf.ps[0])
	}
	ss := make([]string, len(pf.ps))
	for i := range pf.ps {
		ss[i] = PredicateString(pf.ps[i])
	}
	return strings.Join(ss, " ")
}

// Build returns an Any predicate over any collected predicates, or Never if
// none have been collected.
func (pf *PredicateFlag) Build() Predicate {
	if p := Any(pf.ps...); p != nil {
		return p
	}
	return Never
}

// Set implements flag.Value by processing fields from the given string. Each
// field string will be parsed by ParsePredicate.
func (pf *PredicateFlag) Set(s string) error {
	for _, ss := range strings.Fields(s) {
		if err := pf.add(ss); err != nil {
			return err
		}
	}
	return nil
}

// Get implements flag.Getter, it just calls (*PredicateFlag).Build().
func (pf *PredicateFlag) Get() interface{} {
	return pf.Build()
}

func (pf *PredicateFlag) add(s string) error {
	p, err := ParsePredicate(s)
	if err != nil {
		return err
	}
	pf.ps = append(pf.ps, p)
	return nil
}
