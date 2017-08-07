package stackvm

import (
	"bytes"
	"encoding/binary"
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

// New creates a new stack machine with a given program loaded. The prog byte
// array is a sequence of varint encoded unsigned integers (after fixed encoded
// options).
//
// The first fixed byte is a version number, which must currently be 0x00.
//
// The next two bytes encode a 16-bit unsigned stacksize. That much space will
// be reserved in memory for the Parameter Stack (PS) and Control Stack (CS);
// it must be a multiple of the page size.
//
// PS grows up from 0, the PS Base Pointer PBP, to at most stacksize bytes. CS
// grows down from stacksize-1, the CS Base Pointer CBP, towards PS. The
// address of the next slot for PS (resp CS) is stored in the PS Stack Pointer,
// or PSP (resp CSP).
//
// Any push onto either PS or CS will fail with an overflow error when PSP ==
// CSP. Similarly any pop from them will fail with an underflow error when
// their SP meets their BP.
//
// The rest of prog is loaded in memory immediately after the stack space with
// IP pointing at its first byte. Each varint encodes an operation, with the
// lowest 7 bits being the opcode, while all higher bits may encode an
// immediate argument.
//
// For many non-control flow operations, any immediate argument is used in lieu
// of popping a value from the parameter stack. Most control flow operations
// use their immediate argument as an IP offset, however they will consume an
// IP offset from the parameter stack if no immediate is given.
func New(prog []byte) (*Mach, error) {
	p := prog
	if len(p) < 4 {
		return nil, errors.New("program too short, need at least 4 bytes")
	}

	if p[0] != _machVersionCode {
		return nil, fmt.Errorf("unsupported stackvm program version %02x", p[0])
	}
	p = p[1:]

	stackSize := binary.BigEndian.Uint16(p)
	if stackSize%_pageSize != 0 {
		return nil, fmt.Errorf(
			"invalid stacksize %#02x, not a %#02x-multiple",
			stackSize, _pageSize)
	}
	p = p[2:]

	m := Mach{
		ctx: defaultContext,
		opc: makeOpCache(len(p)),
		pbp: 0,
		psp: _pspInit,
		cbp: uint32(stackSize) - 4,
		csp: uint32(stackSize) - 4,
		ip:  uint32(stackSize),
	}

	m.storeBytes(m.ip, p)
	// TODO mark code segment, update data

	return &m, nil
}

func (m *Mach) String() string {
	var buf bytes.Buffer
	buf.WriteString("Mach")
	if m.err != nil {
		if code, halted := m.halted(); halted {
			// TODO: symbolicate
			fmt.Fprintf(&buf, " HALT:%v", code)
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
func (m *Mach) IP() uint32 { return m.ip }

// PBP returns the current parameter stack base pointer.
func (m *Mach) PBP() uint32 { return m.pbp }

// PSP returns the current parameter stack pointer.
func (m *Mach) PSP() uint32 {
	if m.psp > m.cbp {
		return m.pbp
	}
	return m.psp
}

// CBP returns the current control stack base pointer.
func (m *Mach) CBP() uint32 { return m.cbp }

// CSP returns the current control stack pointer.
func (m *Mach) CSP() uint32 { return m.csp }

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
//
// Contextual information may be made available by implementing the Context()
// method: if a tracer wants defines a value for some key, it should return
// that value and a true boolean. Tracers, and other code, may then use
// (*Mach).Tracer().Context() to access contextual information from other
// tracers.
type Tracer interface {
	Context(m *Mach, key string) (interface{}, bool)
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
	return ops[o.Code].name
}

// MachOptions represents options for a machine, currently just stack size (see
// New).
type MachOptions struct {
	StackSize uint16
}

// EncodeInto encodes machine optios for the header of a program.
func (opts MachOptions) EncodeInto(p []byte) int {
	p[0] = _machVersionCode
	binary.BigEndian.PutUint16(p[1:], opts.StackSize)
	return 3
}

// NeededSize returns the number of bytes needed for EncodeInto.
func (opts MachOptions) NeededSize() int { return 3 }

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

// NeededSize returns the number of bytes needed to encode op.
func (o Op) NeededSize() int {
	return int(varOpLength(o.Arg)) + 1
}

// AcceptsRef return true only if the argument can resolve another op reference
// ala ResolveRefArg.
func (o Op) AcceptsRef() bool {
	switch ops[o.Code].imm.kind() {
	case opImmVal, opImmOffset, opImmAddr:
		return true
	}
	return false
}

// ResolveRefArg fills in the argument of a control op relative to another op's
// encoded location, and the current op's.
func (o Op) ResolveRefArg(myIP, targIP uint32) Op {
	switch ops[o.Code].imm.kind() {
	case opImmOffset:
		// need to skip the arg and the code...
		d := targIP - myIP
		n := varOpLength(d)
		d -= n
		if id := int32(d); id < 0 && varOpLength(uint32(id)) != n {
			// ...arg off by one, now that we know its value.
			id--
			d = uint32(id)
		}
		o.Arg = d

	case opImmVal, opImmAddr:
		o.Arg = targIP
	}
	return o
}

func (o Op) String() string {
	def := ops[o.Code]
	if !o.Have {
		return def.name
	}
	switch def.imm.kind() {
	case opImmVal:
		return fmt.Sprintf("%d %s", o.Arg, def.name)
	case opImmAddr:
		return fmt.Sprintf("@%#04x %s", o.Arg, def.name)
	case opImmOffset:
		return fmt.Sprintf("%+#04x %s", o.Arg, def.name)
	}
	return fmt.Sprintf("INVALID(%#x %x %q)", o.Arg, o.Code, def.name)
}

// Tracer returns the current Tracer that the machine is running under, if any.
func (m *Mach) Tracer() Tracer {
	if tc, ok := m.ctx.(tracedContext); ok {
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
func (m *Mach) SetHandler(queueSize int, h Handler) {
	m.ctx = newRunq(h, queueSize)
}

func (tc tracedContext) queue(n *Mach) error {
	tc.t.Queue(tc.m, n)
	n.ctx = tracify(n.ctx, tc.t, n)
	return tc.context.queue(n)
}

// Trace implements the same logic as (*Mach).run, but calls a Tracer
// at the appropriate times.
func (m *Mach) Trace(t Tracer) error {
	// the code below is essentially an
	// instrumented copy of Mach.Run (with mach.run
	// inlined)
	orig := m

	m.ctx = tracify(m.ctx, t, m)

repeat:
	// live
	t.Begin(m)
	for m.err == nil {
		var readOp Op
		if _, code, arg, err := m.read(m.ip); err != nil {
			m.err = err
			break
		} else {
			readOp = Op{code.code(), arg, code.hasImm()}
		}
		t.Before(m, m.ip, readOp)
		m.step()
		if m.err != nil {
			break
		}
		t.After(m, m.ip, readOp)
	}
	t.End(m)

	// win or die
	err := m.ctx.Handle(m)
	t.Handle(m, err)
	if err == nil {
		if n := m.ctx.next(); n != nil {
			m.free()
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

// HaltCode returns the halt code and true if the machine has halted
// normally; otherwise false is returned.
func (m *Mach) HaltCode() (uint32, bool) { return m.halted() }

var (
	lowHaltErrors [256]error
	haltErrors    = make(map[uint32]error)
)

func init() {
	for i := 0; i < len(lowHaltErrors); i++ {
		lowHaltErrors[i] = fmt.Errorf("HALT(%d)", i)
	}
}

// Err returns the last error from machine execution, wrapped with
// execution context.
func (m *Mach) Err() error {
	err := m.err
	if code, halted := m.halted(); halted {
		if code == 0 {
			return nil
		}
		if code < uint32(len(lowHaltErrors)) {
			err = lowHaltErrors[code]
		} else {
			he, def := haltErrors[code]
			if !def {
				he = fmt.Errorf("HALT(%d)", code)
				haltErrors[code] = he
			}
			err = he
		}
	}
	if err == nil {
		return nil
	}
	if _, ok := err.(MachError); !ok {
		return MachError{m.ip, err}
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

func varOpLength(n uint32) (m uint32) {
	for v := n; v != 0; v >>= 7 {
		m++
	}
	m++
	return
}
