package stackvm

type _branchImm uint32
type _bnzImm uint32
type _bzImm uint32

func _branch(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	return m.branch(int32(val))
}

func _bnz(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	if val == 0 {
		return m.cbranch()
	}
	return nil
}

func _bz(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	if val == 0 {
		return m.cbranch()
	}
	return nil
}

func (arg _branchImm) run(m *Mach) error { return m.branch(int32(arg)) }

func (arg _bnzImm) run(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	if val != 0 {
		return m.branch(int32(arg))
	}
	return m.jump(int32(arg))
}

func (arg _bzImm) run(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	if val == 0 {
		return m.branch(int32(arg))
	}
	return m.jump(int32(arg))
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
	return _bnz
}

func bz(arg uint32, have bool) op {
	if have {
		return _bzImm(arg).run
	}
	return _bz
}
