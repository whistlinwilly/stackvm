package stackvm

type _push uint32
type _pop uint32
type _dup uint32
type _swap uint32

func (arg _push) run(m *Mach) error { return m.push(uint32(arg)) }
func (arg _pop) run(m *Mach) error {
	for i := 0; i < int(arg); i++ {
		if err := m.drop(); err != nil {
			return err
		}
	}
	return nil
}

func dup1(m *Mach) error { return m.push(m.pa) }

func (arg _dup) run(m *Mach) error {
	p, err := m.pRef(uint32(arg))
	if err != nil {
		return err
	}
	return m.push(*p)
}

func (arg _swap) run(m *Mach) error {
	if m.psp == _pspInit {
		return stackRangeError{"param", "under"}
	}
	p2, err := m.pRef(1 + uint32(arg))
	if err != nil {
		return err
	}
	m.pa, *p2 = *p2, m.pa
	return nil
}

func push(arg uint32, have bool) op {
	if !have {
		return nil
	}
	return _push(arg).run
}

func pop(arg uint32, have bool) op {
	if !have {
		arg = 1
	}
	return _pop(arg).run
}

func dup(arg uint32, have bool) op {
	if !have || arg == 1 {
		return dup1
	}
	return _dup(arg).run
}

func swap(arg uint32, have bool) op {
	if !have {
		arg = 1
	}
	return _swap(arg).run
}
