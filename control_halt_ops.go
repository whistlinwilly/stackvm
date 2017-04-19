package stackvm

import "fmt"

type _halt uint32
type _hz uint32
type _hnz uint32

func (arg _halt) run(m *Mach) error { return arg }
func (arg _halt) HaltCode() uint32  { return uint32(arg) }
func (arg _halt) Error() string     { return fmt.Sprintf("HALT(%d)", arg) }

func (arg _hz) Error() string    { return fmt.Sprintf("HALT(%d)", arg) }
func (arg _hz) HaltCode() uint32 { return uint32(arg) }
func (arg _hz) run(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	if val == 0 {
		return arg
	}
	return nil
}

func (arg _hnz) Error() string    { return fmt.Sprintf("HALT(%d)", arg) }
func (arg _hnz) HaltCode() uint32 { return uint32(arg) }
func (arg _hnz) run(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	if val != 0 {
		return arg
	}
	return nil
}

func halt(arg uint32, have bool) op { return _halt(arg).run }
func hz(arg uint32, have bool) op   { return _hz(arg).run }
func hnz(arg uint32, have bool) op  { return _hnz(arg).run }
