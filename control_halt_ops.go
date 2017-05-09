package stackvm

import "fmt"

type _halt struct {
	err  error
	code uint32
}

type _hz struct {
	err  error
	code uint32
}

type _hnz struct {
	err  error
	code uint32
}

func (op _halt) run(m *Mach) error { return op.err }
func (op _halt) HaltCode() uint32  { return op.code }
func (op _halt) Error() string     { return fmt.Sprintf("HALT(%d)", op.code) }

func (op _hz) Error() string    { return fmt.Sprintf("HALT(%d)", op.code) }
func (op _hz) HaltCode() uint32 { return op.code }
func (op _hz) run(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	if val == 0 {
		return op.err
	}
	return nil
}

func (op _hnz) Error() string    { return fmt.Sprintf("HALT(%d)", op.code) }
func (op _hnz) HaltCode() uint32 { return op.code }
func (op _hnz) run(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	if val != 0 {
		return op.err
	}
	return nil
}

func halt(arg uint32, have bool) op {
	op := _halt{code: arg}
	op.err = op
	return op.run
}

func hz(arg uint32, have bool) op {
	op := _hz{code: arg}
	op.err = op
	return op.run
}

func hnz(arg uint32, have bool) op {
	op := _hnz{code: arg}
	op.err = op
	return op.run
}
