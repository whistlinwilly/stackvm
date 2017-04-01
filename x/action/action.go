package action

import (
	"errors"
	"fmt"

	"github.com/jcorbin/stackvm"
)

var errInvalidTraceAction = errors.New("invalid trace action")

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

// Test returns true if the current trace action is the received one.
func (ta TraceAction) Test(act TraceAction, _ uint32, _ stackvm.Op) bool { return act == ta }

func (ta TraceAction) String() string {
	switch ta {
	case TraceBegin:
		return "begin"
	case TraceEnd:
		return "end"
	case TraceQueue:
		return "queue"
	case TraceBefore:
		return "before"
	case TraceAfter:
		return "after"
	default:
		return fmt.Sprintf("InvalidTraceAction(%d)", int(ta))
	}
}

// Get returns the trace action value for the flag.Getter interface.
func (ta *TraceAction) Get() interface{} { return *ta }

// Set sets the trace action value from a flag string.
func (ta *TraceAction) Set(s string) error {
	switch s {
	case "begin":
		*ta = TraceBegin
	case "end":
		*ta = TraceEnd
	case "queue":
		*ta = TraceQueue
	case "before":
		*ta = TraceBefore
	case "after":
		*ta = TraceAfter
	default:
		return errInvalidTraceAction
	}
	return nil
}
