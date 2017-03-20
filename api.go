package stackvm

import "fmt"

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
	addr int
	err  error
}

// Cause returns the underlying machine error.
func (me MachError) Cause() error { return me.err }

func (me MachError) Error() string { return fmt.Sprintf("@%04x: %v", me.addr, me.err) }
