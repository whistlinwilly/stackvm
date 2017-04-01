package action

import "github.com/jcorbin/stackvm"

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
