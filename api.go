package stackvm

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
)

// New creates a new stack machine. At least stackSize bytes are
// reserved for the parameter and control stacks (combined, they
// grow towards each other); the actual amount reserved is rounded
// up to whole memory pages.
func New(stackSize uint32) *Mach {
	stackSize += stackSize % _pageSize
	pbp := uint32(_stackBase)
	cbp := pbp + stackSize - 4
	return &Mach{
		pbp: pbp,
		psp: pbp,
		cbp: cbp,
		csp: cbp,
	}
}

func (m *Mach) String() string {
	var buf bytes.Buffer
	buf.WriteString("Mach")
	if m.err != nil {
		// TODO: symbolicate
		fmt.Fprintf(&buf, " ERR:%v", m.err)
	}
	fmt.Fprintf(&buf, " @0x%04x 0x%04x:0x%04x 0x%04x:0x%04x", m.ip, m.pbp, m.psp, m.cbp, m.csp)
	// TODO:
	// pages?
	// stack dump?
	// context describe?
	return buf.String()
}

// Load loads machine code into memory, and sets IP to point at the
// beginning of the loaded bytes.
func (m *Mach) Load(prog []byte) error {
	m.ip = m.cbp + 1
	m.ip += m.ip % _pageSize
	m.storeBytes(m.ip, prog)
	// TODO mark code segment, update data
	return nil
}

// Dump hex dumps all machine memory to a given io.Writer.
func (m *Mach) Dump(w io.Writer) (err error) {
	var z [_pageSize]byte
	d := hex.Dumper(w)
	for _, pg := range m.pages {
		if pg != nil {
			_, err = d.Write(pg.d[:])
		} else {
			_, err = d.Write(z[:])
		}
		if err != nil {
			break
		}
	}
	return err
}

// IP returns the current instruction pointer.
func (m *Mach) IP() uint32 {
	return m.ip
}

// Stacks returns the current values on the parameter and control
// stacks.
func (m *Mach) Stacks() (ps, cs []uint32, err error) {
	var val uint32
	csp := m.csp
	psp := m.psp
	if psp > m.csp {
		psp = m.csp
	}
	if psp > m.cbp {
		psp = m.cbp
	}
	if csp > m.csp {
		csp = m.csp
	}
	if csp > m.cbp {
		csp = m.cbp
	}
	for addr := m.pbp; addr < psp; addr += 4 {
		val, err = m.fetch(addr)
		if err != nil {
			return
		}
		ps = append(ps, val)
	}
	for addr := m.cbp; addr > csp; addr -= 4 {
		val, err = m.fetch(addr)
		if err != nil {
			return
		}
		cs = append(cs, val)
	}
	return
}

// MemCopy copies bytes from memory into the given buffer, returning
// the number of bytes copied.
func (m *Mach) MemCopy(addr uint32, bs []byte) int {
	return m.fetchBytes(addr, bs)
}

// Tracer is the interface taken by (*Mach).Trace to observe machine
// execution.
type Tracer interface {
	Begin(m *Mach)
	Before(m *Mach, ip uint32, op Op)
	After(m *Mach, ip uint32, op Op)
	Queue(m, n *Mach)
	End(m *Mach)
	Handle(m *Mach, err error)
}

// Op is used within Tracer to pass along decoded machine operations.
type Op struct {
	Code byte
	Arg  uint32
	Have bool
}

func (o Op) String() string {
	if !o.Have {
		return opCode2Name[o.Code]
	}
	// TODO: better formatting for ip offset immediates
	return fmt.Sprintf("%d %s", o.Arg, opCode2Name[o.Code])
}

type tracedContext struct {
	context
	t Tracer
	m *Mach
}

// SetHandler allocates a pending queue and sets a result handling
// function. Without a pending queue, the fork family of operations
// will fail. Without a result handling function, there's not much
// point to running more than one machine.
func (m *Mach) SetHandler(queueSize int, f func(*Mach) error) {
	m.ctx = newRunq(handler(f), queueSize)
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
			t.Before(m, m.ip, Op{code, arg, have})
			op, err := makeOp(code, arg, have)
			if err != nil {
				m.err = err
				break
			}
			m.ip = ip
			t.After(m, m.ip, Op{code, arg, have})
			if err := op(m); err != nil {
				m.err = err
				break
			}
		}
		t.End(m)
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

func (me MachError) Error() string { return fmt.Sprintf("@0x%04x: %v", me.addr, me.err) }
