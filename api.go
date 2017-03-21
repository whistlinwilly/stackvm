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

// Run runs the machine until termination, returning any error.
func (m *Mach) Run() error {
	m.run()
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
	if err != nil {
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
