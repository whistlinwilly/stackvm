package stackvm

type _push uint32
type _pop uint32
type _dup uint32
type _swap uint32

func (arg _push) run(m *Mach) error { return m.push(uint32(arg)) }
func (arg _pop) run(m *Mach) error {
	for i := 0; i < int(arg); i++ {
		_, err := m.pop()
		if err != nil {
			return err
		}
	}
	return nil
}

func (arg _dup) run(m *Mach) error {
	addr, err := m.pAddr(int32(arg))
	if err != nil {
		return err
	}
	val, err := m.fetch(addr)
	if err != nil {
		return err
	}
	return m.push(val)
}

func (arg _swap) run(m *Mach) error {
	addr1, err := m.pAddr(0)
	if err != nil {
		return err
	}
	addr2, err := m.pAddr(int32(arg))
	if err != nil {
		return err
	}
	val1, err := m.fetch(addr1)
	if err != nil {
		return err
	}
	val2, err := m.fetch(addr2)
	if err != nil {
		return err
	}
	if err := m.store(addr1, val2); err != nil {
		return err
	}
	return m.store(addr2, val1)
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
	return _dup(arg).run
}

func swap(arg uint32, have bool) op {
	return _swap(arg).run
}
