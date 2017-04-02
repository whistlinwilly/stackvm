package stackvm

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

var errRunning = errors.New("machine running")

// NoSuchOpError is returned by ResolveOp if the named operation is not //
// defined.
type NoSuchOpError string

func (name NoSuchOpError) Error() string {
	return fmt.Sprintf("no such operation %q", string(name))
}

// New creates a new stack machine. At least stackSize bytes are
// reserved for the parameter and control stacks (combined, they
// grow towards each other); the actual amount reserved is rounded
// up to whole memory pages.
func New(stackSize uint32) *Mach {
	stackSize += stackSize % _pageSize
	pbp := uint32(_stackBase)
	cbp := pbp + stackSize - 4
	return &Mach{
		context: defaultContext,
		pbp:     pbp,
		psp:     pbp,
		cbp:     cbp,
		csp:     cbp,
	}
}

func (m *Mach) String() string {
	var buf bytes.Buffer
	buf.WriteString("Mach")
	if m.err != nil {
		if arg, ok := m.err.(_halt); ok {
			// TODO: symbolicate
			fmt.Fprintf(&buf, " HALT:%v", int(arg))
		} else {
			fmt.Fprintf(&buf, " ERR:%v", m.err)
		}
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
	m.ip = m.cbp + 4
	m.ip += m.ip % _pageSize
	m.storeBytes(m.ip, prog)
	// TODO mark code segment, update data
	return nil
}

// EachPage calls a function with each allocated section of memory; it MUST NOT
// mutate the memory, and should copy out any data that it needs to retain.
func (m *Mach) EachPage(f func(addr uint32, p [64]byte) error) error {
	for i, pg := range m.pages {
		if pg != nil {
			if err := f(uint32(i*_pageSize), pg.d); err != nil {
				return err
			}
		}
	}
	return nil
}

var zeroPageData [_pageSize]byte

// WriteTo writes all machine memory to the given io.Writer, returning the
// number of bytes written.
func (m *Mach) WriteTo(w io.Writer) (n int64, err error) {
	for _, pg := range m.pages {
		var wn int
		if pg == nil {
			wn, err = w.Write(zeroPageData[:])
		} else {
			wn, err = w.Write(pg.d[:])
		}
		n += int64(wn)
		if err != nil {
			break
		}
	}
	return
}

// IP returns the current instruction pointer.
func (m *Mach) IP() uint32 {
	return m.ip
}

// PBP returns the current parameter stack base pointer.
func (m *Mach) PBP() uint32 {
	return m.pbp
}

// PSP returns the current parameter stack pointer.
func (m *Mach) PSP() uint32 {
	return m.psp
}

// CBP returns the current control stack base pointer.
func (m *Mach) CBP() uint32 {
	return m.cbp
}

// CSP returns the current control stack pointer.
func (m *Mach) CSP() uint32 {
	return m.csp
}

// Values returns any recorded result values from a finished machine. After a
// machine halts with 0 status code, the control stack may contain zero or
// more pairs of memory address ranges. If so, then Values will extract all
// such ranged values, and return them as a slice-of-slices.
func (m *Mach) Values() ([][]uint32, error) {
	if m.err == nil {
		return nil, errRunning
	}

	if arg, ok := m.halted(); !ok || arg != 0 {
		if m.err != nil {
			return nil, m.err
		}
		return nil, errRunning
	}

	cs, err := m.fetchCS()
	if err != nil {
		return nil, err
	}
	if len(cs)%2 != 0 {
		return nil, fmt.Errorf("invalid control stack length %d", len(cs))
	}
	if len(cs) == 0 {
		return nil, nil
	}

	res := make([][]uint32, 0, len(cs)/2)
	for i := 0; i < len(cs); i += 2 {
		ns, err := m.fetchMany(cs[i], cs[i+1])
		if err != nil {
			return nil, err
		}
		res = append(res, ns)
	}
	return res, nil
}

// Stacks returns the current values on the parameter and control
// stacks.
func (m *Mach) Stacks() ([]uint32, []uint32, error) {
	ps, err := m.fetchPS()
	if err != nil {
		return nil, nil, err
	}
	cs, err := m.fetchCS()
	if err != nil {
		return nil, nil, err
	}
	return ps, cs, nil
}

// MemCopy copies bytes from memory into the given buffer, returning
// the number of bytes copied.
func (m *Mach) MemCopy(addr uint32, bs []byte) int {
	return m.fetchBytes(addr, bs)
}

// Tracer is the interface taken by (*Mach).Trace to observe machine
// execution: Begin() and End() are called when a machine starts and finishes
// respectively; Before() and After() are around each machine operation;
// Queue() is called when a machine creates a copy of itself; Handle() is
// called after an ended machine has been passed to any result handling
// function.
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

// ResolveOp builds an op given a name string, and argument.
func ResolveOp(name string, arg uint32, have bool) (Op, error) {
	code, def := opName2Code[name]
	if !def {
		return Op{}, NoSuchOpError(name)
	}
	return Op{code, arg, have}, nil
}

// Name returns the name of the coded operation.
func (o Op) Name() string {
	return opCode2Name[o.Code]
}

// EncodeInto encodes the operation into the given buffer, returning the number
// of bytes encoded.
func (o Op) EncodeInto(p []byte) int {
	var ep [6]byte
	k, i := 0, 6
	i--
	ep[i] = o.Code
	if o.Have {
		v := o.Arg
		for {
			i--
			if i < 0 {
				break
			}
			ep[i] = byte(v) | 0x80
			v >>= 7
			if v == 0 {
				break
			}
		}
	}
	for i < len(ep) && k < len(p) {
		p[k] = ep[i]
		i++
		k++
	}
	return k
}

func (o Op) String() string {
	if !o.Have {
		return opCode2Name[o.Code]
	}
	// TODO: better formatting for ip offset immediates
	return fmt.Sprintf("%d %s", o.Arg, opCode2Name[o.Code])
}

// Tracer returns the current Tracer that the machine is running under, if any.
func (m *Mach) Tracer() Tracer {
	if tc, ok := m.context.(tracedContext); ok {
		return tc.t
	}
	return nil
}

type tracedContext struct {
	context
	t Tracer
	m *Mach
}

func tracify(ctx context, t Tracer, m *Mach) context {
	for tc, ok := ctx.(tracedContext); ok; tc, ok = ctx.(tracedContext) {
		ctx = tc.context
	}
	return tracedContext{ctx, t, m}
}

// SetHandler allocates a pending queue and sets a result handling
// function. Without a pending queue, the fork family of operations
// will fail. Without a result handling function, there's not much
// point to running more than one machine.
func (m *Mach) SetHandler(queueSize int, f func(*Mach) error) {
	m.context = newRunq(handler(f), queueSize)
}

func (tc tracedContext) queue(n *Mach) error {
	tc.t.Queue(tc.m, n)
	n.context = tracify(n.context, tc.t, n)
	return tc.context.queue(n)
}

// Trace implements the same logic as (*Mach).run, but calls a Tracer
// at the appropriate times.
func (m *Mach) Trace(t Tracer) error {
	// the code below is essentially an
	// instrumented copy of Mach.Run (with mach.run
	// inlined)
	orig := m

	if m.context != defaultContext {
		m.context = tracify(m.context, t, m)
	}

repeat:
	// live
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
		if err := op(m); err != nil {
			m.err = err
			break
		}
		t.After(m, m.ip, Op{code, arg, have})
	}
	t.End(m)

	// win or die
	err := m.handle(m)
	t.Handle(m, err)
	if err == nil {
		if n := m.next(); n != nil {
			m = n
			// die
			goto repeat
		}
	}

	// win?
	if m != orig {
		*orig = *m
	}
	return err
}

// Run runs the machine until termination, returning any error.
func (m *Mach) Run() error {
	n, err := m.run()
	if n != m {
		*m = *n
	}
	return err
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

	if code, ok := m.halted(); ok && code == 0 {
		err = nil
	}
	// TODO: provide non-zero halt error table

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
