package stackvm

type _push uint32
type _pop uint32
type _dup uint32
type _swap uint32

func (arg _push) run(m *Mach) error { return m.push(byte(arg)) }
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
	return m.push(m.fetch(addr))
}

func (arg _swap) run(m *Mach) error {
	addr1, err := m.pAddr(int32(arg))
	if err != nil {
		return err
	}
	addr2, err := m.pAddr(int32(arg + 1))
	if err != nil {
		return err
	}
	// TODO: better with gather -> scatter
	i1, _, pg1 := m.pageFor(addr1)
	i2, _, pg2 := m.pageFor(addr2)
	pg1.d[i1], pg2.d[i2] = pg2.d[i2], pg1.d[i1]
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
	return _dup(arg).run
}

func swap(arg uint32, have bool) op {
	return _swap(arg).run
}
