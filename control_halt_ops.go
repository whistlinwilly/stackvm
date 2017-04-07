package stackvm

import "fmt"

type _halt uint32

func (arg _halt) run(m *Mach) error { return arg }

func (arg _halt) Error() string {
	return fmt.Sprintf("HALT(%d)", arg)
}

func halt(arg uint32, have bool) op {
	return _halt(arg).run
}
