package stackvm

import "fmt"

// New creates a new stack machine. At least stackSize bytes are
// reserved for the parameter and control stacks (combined, they
// grow towards each other); the actual amount reserved is rounded
// up to whole memory pages.
func New(stackSize uint32) *Mach {
	stackSize += stackSize % _pageSize
	return &Mach{
		pbp: _stackBase,
		cbp: _stackBase + stackSize - 1,
	}
}

// Load loads machine code into memory, and sets IP to point at the
// beginning of the loaded bytes.
func (m *Mach) Load(prog []byte) error {
	m.ip = m.cbp + 1
	m.ip += m.ip % _pageSize
	m.storeBytes(m.ip, prog)
	return nil
}

// Tracer is the interface taken by (*Mach).Trace to observe machine
// execution.
type Tracer interface {
	Begin(m *Mach)
	Before(m *Mach, ip uint32, op Op)
	After(m *Mach, ip uint32, op Op)
	Queue(m, n *Mach)
	End(m *Mach, err error)
	Handle(m *Mach, err error)
}

// Op is used within Tracer to pass along decoded machine operations.
type Op struct {
	Code byte
	Arg  uint32
}

type tracedContext struct {
	context
	t Tracer
	m *Mach
}

func (tc tracedContext) queue(n *Mach) error {
	tc.t.Queue(tc.m, n)
	n.ctx = tracedContext{n.ctx, tc.t, n}
	return tc.context.queue(n)
}

// Trace implements the same logic as (*Mach).run, but calls a Tracer
// at the appropriate times.
func (m *Mach) Trace(t Tracer) error {
	orig := m

	if m.ctx != nil {
		m.ctx = tracedContext{m.ctx, t, m}
	}

	for m.err == nil {
		t.Begin(m)
		for m.err == nil {
			ip, code, arg, have, err := m.decode(m.ip)
			if err != nil {
				m.err = err
				break
			}
			t.Before(m, m.ip, Op{code, arg})
			op, err := makeOp(code, arg, have)
			if err != nil {
				m.err = err
				break
			}
			m.ip = ip
			t.After(m, m.ip, Op{code, arg})
			if err := op(m); err != nil {
				m.err = err
				break
			}
		}
		t.End(m, m.Err())
		if m.ctx != nil {
			m.err = m.ctx.handle(m)
			t.Handle(m, m.err)
		}
		if m.ctx != nil && m.err == nil {
			if n := m.ctx.next(); n != nil {
				m = n
			}
		}
	}
	if m != orig {
		*orig = *m
	}
	return m.Err()
}

// Run runs the machine until termination, returning any error.
func (m *Mach) Run() error {
	n := m.run()
	if n != m {
		*m = *n
	}
	return m.Err()
}

// Step single steps the machine; it decodes and executes one
// operation.
func (m *Mach) Step() error {
	if m.err == nil {
		m.step()
	}
	return m.Err()
}

// Err returns the last error from machine execution, wrapped with
// execution context.
func (m *Mach) Err() error {
	err := m.err
	if arg, ok := err.(_halt); ok {
		if arg == 0 {
			err = nil
		}
		// TODO: provide non-zero halt error table
	}
	if _, ok := err.(MachError); !ok && err != nil {
		err = MachError{m.ip, err}
	}
	return err
}

// MachError wraps an underlying machine error with machine state.
type MachError struct {
	addr uint32
	err  error
}

// Cause returns the underlying machine error.
func (me MachError) Cause() error { return me.err }

func (me MachError) Error() string { return fmt.Sprintf("@%04x: %v", me.addr, me.err) }
