package action

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// TODO: maybe support
// - IP ranges
// - explicit "and"/"or"
// - explicit "not"
// - big feature: matching on trace id (currently an internal detail of LogfTracer)

var errNotAPredicate = errors.New("not a predicate string")

var predicateParsePattern = regexp.MustCompile(stripSpace(`
	^
	(?P<action> \w+ )?
	(?:
		@ (?: (?P<ipDec> [0-9]+ ) | 0x(?P<ipHex> [0-9a-fA-F]+ ) ) |
		: (?P<op> \w+ (?: , \w+ )* )
	)?
	$
`))

// ParsePredicate parses a string like "action", "action@ip", or
// "action:op[,op[,...]" into a predicate. The action string may be any of
// "begin", "end", "queue", "before", or "after" (corresponding to Tracer
// methods). The ip string may either be a decimal number, or a "0x" prefixed
// hex number. The op strings may be any valid operation names.
func ParsePredicate(s string) (Predicate, error) {
	if s == "" {
		return Never, nil
	}

	parts := predicateParsePattern.FindStringSubmatch(s)
	if parts == nil {
		return nil, errNotAPredicate
	}

	ps := make([]Predicate, 0, 2)

	if parts[1] != "" { // group 1 action
		var act TraceAction
		if err := act.Set(parts[1]); err != nil {
			return nil, err
		}
		ps = append(ps, act)
	}

	if parts[2] != "" { // group 2 ipDec
		ip, err := strconv.ParseUint(parts[2], 10, 32)
		if err != nil {
			return nil, err
		}
		ps = append(ps, isIP(ip))
	} else if parts[3] != "" { // group 3 ipHex
		ip, err := strconv.ParseUint(parts[3], 16, 32)
		if err != nil {
			return nil, err
		}
		ps = append(ps, isIP(ip))
	} else if parts[4] != "" { // group 4 op
		ops := strings.Split(parts[4], ",")
		// TODO: validate op names, maybe resolve to codes
		if len(ops) == 1 {
			ps = append(ps, isOPName(ops[0]))
		} else {
			ps = append(ps, anyOPName(ops))
		}
	}

	if p := All(ps...); p != nil {
		return p, nil
	}
	return Never, nil
}

// PredicateString returns a string parse-able by ParsePredicate, or
// "!TYPE(VAL)" if the given predicate wouldn't have been possible from
// ParsePredicate.
func PredicateString(p Predicate) string {
	if p == Never {
		return ""
	}
	switch v := p.(type) {
	case allPredicate:
		if len(v) != 2 {
			break
		}
		if ta, ok := v[0].(TraceAction); ok {
			if sub := PredicateString(v[1]); sub != "" {
				return fmt.Sprintf("%v%s", ta, sub)
			}
		}
	case TraceAction:
		return v.String()
	case isIP:
		return fmt.Sprintf("@%#x", v)
	case isOPName:
		return ":" + string(v)
	case anyOPName:
		return ":" + strings.Join(v, ",")
	}
	return fmt.Sprintf("!%T(%#v)", p, p)
}

func stripSpace(s string) string {
	s = strings.Replace(s, "\n", "", -1)
	s = strings.Replace(s, "\t", "", -1)
	s = strings.Replace(s, " ", "", -1)
	return s
}
