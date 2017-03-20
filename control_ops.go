package stackvm

import "fmt"

type _jumpImm uint32
type _jnzImm uint32
type _jzImm uint32
type _forkImm uint32
type _fnzImm uint32
type _fzImm uint32
type _branchImm uint32
type _bnzImm uint32
type _bzImm uint32
type _callImm uint32

func _jump(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	return m.jump(int(val))
}

func _fork(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	return m.fork(int(val))
}

func _branch(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	return m.branch(int(val))
}

func _ret(m *Mach) error { return m.ret() }
func _call(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	return m.call(int(val))
}

func (arg _jumpImm) run(m *Mach) error   { return m.jump(int(arg)) }
func (arg _forkImm) run(m *Mach) error   { return m.fork(int(arg)) }
func (arg _branchImm) run(m *Mach) error { return m.branch(int(arg)) }
func (arg _callImm) run(m *Mach) error   { return m.call(int(arg)) }

func (arg _jnzImm) run(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	if val != 0 {
		return m.jump(int(arg))
	}
	return nil
}

func (arg _fnzImm) run(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	if val != 0 {
		return m.fork(int(arg))
	}
	return nil
}

func (arg _bnzImm) run(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	if val != 0 {
		return m.branch(int(arg))
	}
	return nil
}

func (arg _jzImm) run(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	if val == 0 {
		return m.jump(int(arg))
	}
	return nil
}

func (arg _fzImm) run(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	if val == 0 {
		return m.fork(int(arg))
	}
	return nil
}

func (arg _bzImm) run(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	if val == 0 {
		return m.branch(int(arg))
	}
	return nil
}

func jump(arg uint32, have bool) op {
	if have {
		return _jumpImm(arg).run
	}
	return _jump
}

func jnz(arg uint32, have bool) op {
	if have {
		return _jnzImm(arg).run
	}
	return nil
}
func jz(arg uint32, have bool) op {
	if have {
		return _jzImm(arg).run
	}
	return nil
}

func fork(arg uint32, have bool) op {
	if have {
		return _forkImm(arg).run
	}
	return _fork
}

func fnz(arg uint32, have bool) op {
	if have {
		return _fnzImm(arg).run
	}
	return nil
}

func fz(arg uint32, have bool) op {
	if have {
		return _fzImm(arg).run
	}
	return nil
}

func branch(arg uint32, have bool) op {
	if have {
		return _branchImm(arg).run
	}
	return _branch
}

func bnz(arg uint32, have bool) op {
	if have {
		return _bnzImm(arg).run
	}
	return nil
}

func bz(arg uint32, have bool) op {
	if have {
		return _bzImm(arg).run
	}
	return nil
}

func call(arg uint32, have bool) op {
	if have {
		return _callImm(arg).run
	}
	return _call
}

func ret(arg uint32, have bool) op {
	if have {
		return nil
	}
	return _ret
}

type _cpop uint32
type _p2c uint32
type _c2p uint32

func (arg _cpop) run(m *Mach) error {
	for i := 0; i < int(arg); i++ {
		_, err := m.cpop()
		if err != nil {
			return err
		}
	}
	return nil
}

func (arg _p2c) run(m *Mach) error {
	for i := 0; i < int(arg); i++ {
		val, err := m.pop()
		if err != nil {
			return err
		}
		if err := m.cpush(val); err != nil {
			return err
		}
	}
	return nil
}

func (arg _c2p) run(m *Mach) error {
	for i := 0; i < int(arg); i++ {
		val, err := m.cpop()
		if err != nil {
			return err
		}
		if err := m.push(val); err != nil {
			return err
		}
	}
	return nil
}

func cpop(arg uint32, have bool) op {
	if !have {
		arg = 1
	}
	return _cpop(arg).run
}

func p2c(arg uint32, have bool) op {
	if !have {
		arg = 1
	}
	return _p2c(arg).run
}

func c2p(arg uint32, have bool) op {
	if !have {
		arg = 1
	}
	return _c2p(arg).run
}

type _halt uint32

func (arg _halt) run(m *Mach) error { return arg }

func (arg _halt) Error() string {
	return fmt.Sprintf("HALT(%d)", arg)
}

func halt(arg uint32, have bool) op {
	return _halt(arg).run
}