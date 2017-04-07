package stackvm

type _forkImm uint32
type _fnzImm uint32
type _fzImm uint32

func _fork(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	return m.fork(int32(val))
}

func _fnz(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	if val != 0 {
		return m.cfork()
	}
	return nil
}

func _fz(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	if val == 0 {
		return m.cfork()
	}
	return nil
}

func (arg _forkImm) run(m *Mach) error { return m.fork(int32(arg)) }

func (arg _fnzImm) run(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	if val != 0 {
		return m.fork(int32(arg))
	}
	return nil
}

func (arg _fzImm) run(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	if val == 0 {
		return m.fork(int32(arg))
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
	return _fnz
}

func fz(arg uint32, have bool) op {
	if have {
		return _fzImm(arg).run
	}
	return _fz
}
