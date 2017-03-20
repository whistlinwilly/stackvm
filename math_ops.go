package stackvm

type _neg uint32
type _addImm uint32
type _subImm uint32
type _mulImm uint32
type _divImm uint32
type _modImm uint32
type _divmodImm uint32

func (arg _neg) run(m *Mach) error {
	addr, err := m.pAddr(int(arg))
	if err != nil {
		return err
	}
	i, _, pg := m.pageFor(addr)
	pg.d[i] = -pg.d[i]
	return nil
}

func _add(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	addr, err := m.pAddr(1)
	if err != nil {
		return err
	}
	i, _, pg := m.pageFor(addr)
	pg.d[i] += val
	return nil
}

func _sub(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	addr, err := m.pAddr(1)
	if err != nil {
		return err
	}
	i, _, pg := m.pageFor(addr)
	pg.d[i] -= val
	return nil
}

func _mul(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	addr, err := m.pAddr(1)
	if err != nil {
		return err
	}
	i, _, pg := m.pageFor(addr)
	pg.d[i] *= val
	return nil
}

func _div(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	addr, err := m.pAddr(1)
	if err != nil {
		return err
	}
	i, _, pg := m.pageFor(addr)
	pg.d[i] /= val
	return nil
}

func _mod(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	addr, err := m.pAddr(1)
	if err != nil {
		return err
	}
	i, _, pg := m.pageFor(addr)
	pg.d[i] %= val
	return nil
}

func _divmod(m *Mach) error {
	addr1, err := m.pAddr(0)
	if err != nil {
		return err
	}
	addr2, err := m.pAddr(1)
	if err != nil {
		return err
	}
	i1, _, pg1 := m.pageFor(addr1)
	i2, _, pg2 := m.pageFor(addr2)
	a, b := pg1.d[i1], pg2.d[i2]
	pg1.d[i1], pg2.d[i2] = a/b, a%b
	return nil
}

func (arg _addImm) run(m *Mach) error {
	addr, err := m.pAddr(0)
	if err != nil {
		return err
	}
	i, _, pg := m.pageFor(addr)
	pg.d[i] += byte(arg)
	return nil
}

func (arg _subImm) run(m *Mach) error {
	addr, err := m.pAddr(0)
	if err != nil {
		return err
	}
	i, _, pg := m.pageFor(addr)
	pg.d[i] -= byte(arg)
	return nil
}

func (arg _mulImm) run(m *Mach) error {
	addr, err := m.pAddr(0)
	if err != nil {
		return err
	}
	i, _, pg := m.pageFor(addr)
	pg.d[i] *= byte(arg)
	return nil
}

func (arg _divImm) run(m *Mach) error {
	addr, err := m.pAddr(0)
	if err != nil {
		return err
	}
	i, _, pg := m.pageFor(addr)
	pg.d[i] /= byte(arg)
	return nil
}

func (arg _modImm) run(m *Mach) error {
	addr, err := m.pAddr(0)
	if err != nil {
		return err
	}
	i, _, pg := m.pageFor(addr)
	pg.d[i] %= byte(arg)
	return nil
}

func (arg _divmodImm) run(m *Mach) error {
	addr, err := m.pAddr(0)
	if err != nil {
		return err
	}
	i, _, pg := m.pageFor(addr)
	a := pg.d[i]
	pg.d[i] = a / byte(arg)
	return m.push(a % byte(arg))
}

func neg(arg uint32, have bool) op {
	return _neg(arg).run
}

func add(arg uint32, have bool) op {
	if !have {
		return _add
	}
	return _addImm(arg).run
}

func sub(arg uint32, have bool) op {
	if !have {
		return _sub
	}
	return _subImm(arg).run
}

func mul(arg uint32, have bool) op {
	if !have {
		return _mul
	}
	return _mulImm(arg).run
}

func div(arg uint32, have bool) op {
	if !have {
		return _div
	}
	return _divImm(arg).run
}

func mod(arg uint32, have bool) op {
	if !have {
		return _mod
	}
	return _modImm(arg).run
}

func divmod(arg uint32, have bool) op {
	if !have {
		return _divmod
	}
	return _divmodImm(arg).run
}
