package stackvm

type _ltImm uint32
type _lteImm uint32
type _eqImm uint32
type _neqImm uint32
type _gtImm uint32
type _gteImm uint32

func _lt(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	addr, err := m.pAddr(1)
	if err != nil {
		return err
	}
	i, _, pg := m.pageFor(addr)
	if pg.d[i] < val {
		pg.d[i] = 1
	} else {
		pg.d[i] = 0
	}
	return nil
}

func _lte(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	addr, err := m.pAddr(1)
	if err != nil {
		return err
	}
	i, _, pg := m.pageFor(addr)
	if pg.d[i] <= val {
		pg.d[i] = 1
	} else {
		pg.d[i] = 0
	}
	return nil
}

func _eq(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	addr, err := m.pAddr(1)
	if err != nil {
		return err
	}
	i, _, pg := m.pageFor(addr)
	if pg.d[i] == val {
		pg.d[i] = 1
	} else {
		pg.d[i] = 0
	}
	return nil
}

func _neq(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	addr, err := m.pAddr(1)
	if err != nil {
		return err
	}
	i, _, pg := m.pageFor(addr)
	if pg.d[i] != val {
		pg.d[i] = 1
	} else {
		pg.d[i] = 0
	}
	return nil
}

func _gt(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	addr, err := m.pAddr(1)
	if err != nil {
		return err
	}
	i, _, pg := m.pageFor(addr)
	if pg.d[i] > val {
		pg.d[i] = 1
	} else {
		pg.d[i] = 0
	}
	return nil
}

func _gte(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	addr, err := m.pAddr(1)
	if err != nil {
		return err
	}
	i, _, pg := m.pageFor(addr)
	if pg.d[i] >= val {
		pg.d[i] = 1
	} else {
		pg.d[i] = 0
	}
	return nil
}

func (arg _ltImm) run(m *Mach) error {
	addr, err := m.pAddr(0)
	if err != nil {
		return err
	}
	i, _, pg := m.pageFor(addr)
	if pg.d[i] < byte(arg) {
		pg.d[i] = 1
	} else {
		pg.d[i] = 0
	}
	return nil
}

func (arg _lteImm) run(m *Mach) error {
	addr, err := m.pAddr(0)
	if err != nil {
		return err
	}
	i, _, pg := m.pageFor(addr)
	if pg.d[i] <= byte(arg) {
		pg.d[i] = 1
	} else {
		pg.d[i] = 0
	}
	return nil
}

func (arg _eqImm) run(m *Mach) error {
	addr, err := m.pAddr(0)
	if err != nil {
		return err
	}
	i, _, pg := m.pageFor(addr)
	if pg.d[i] == byte(arg) {
		pg.d[i] = 1
	} else {
		pg.d[i] = 0
	}
	return nil
}

func (arg _neqImm) run(m *Mach) error {
	addr, err := m.pAddr(0)
	if err != nil {
		return err
	}
	i, _, pg := m.pageFor(addr)
	if pg.d[i] != byte(arg) {
		pg.d[i] = 1
	} else {
		pg.d[i] = 0
	}
	return nil
}

func (arg _gtImm) run(m *Mach) error {
	addr, err := m.pAddr(0)
	if err != nil {
		return err
	}
	i, _, pg := m.pageFor(addr)
	if pg.d[i] > byte(arg) {
		pg.d[i] = 1
	} else {
		pg.d[i] = 0
	}
	return nil
}

func (arg _gteImm) run(m *Mach) error {
	addr, err := m.pAddr(0)
	if err != nil {
		return err
	}
	i, _, pg := m.pageFor(addr)
	if pg.d[i] >= byte(arg) {
		pg.d[i] = 1
	} else {
		pg.d[i] = 0
	}
	return nil
}

func lt(arg uint32, have bool) op {
	if !have {
		return _lt
	}
	return _ltImm(arg).run
}

func lte(arg uint32, have bool) op {
	if !have {
		return _lte
	}
	return _lteImm(arg).run
}

func eq(arg uint32, have bool) op {
	if !have {
		return _eq
	}
	return _eqImm(arg).run
}

func neq(arg uint32, have bool) op {
	if !have {
		return _neq
	}
	return _neqImm(arg).run
}

func gt(arg uint32, have bool) op {
	if !have {
		return _gt
	}
	return _gtImm(arg).run
}

func gte(arg uint32, have bool) op {
	if !have {
		return _gte
	}
	return _gteImm(arg).run
}
