package stackvm

type _ltImm uint32
type _lteImm uint32
type _eqImm uint32
type _neqImm uint32
type _gtImm uint32
type _gteImm uint32

func _lt(m *Mach) error {
	b, err := m.pop()
	if err != nil {
		return err
	}
	a, err := m.pop()
	if err != nil {
		return err
	}
	if a < b {
		return m.push(1)
	}
	return m.push(0)
}

func _lte(m *Mach) error {
	b, err := m.pop()
	if err != nil {
		return err
	}
	a, err := m.pop()
	if err != nil {
		return err
	}
	if a <= b {
		return m.push(1)
	}
	return m.push(0)
}

func _eq(m *Mach) error {
	b, err := m.pop()
	if err != nil {
		return err
	}
	a, err := m.pop()
	if err != nil {
		return err
	}
	if a == b {
		return m.push(1)
	}
	return m.push(0)
}

func _neq(m *Mach) error {
	b, err := m.pop()
	if err != nil {
		return err
	}
	a, err := m.pop()
	if err != nil {
		return err
	}
	if a != b {
		return m.push(1)
	}
	return m.push(0)
}

func _gt(m *Mach) error {
	b, err := m.pop()
	if err != nil {
		return err
	}
	a, err := m.pop()
	if err != nil {
		return err
	}
	if a > b {
		return m.push(1)
	}
	return m.push(0)
}

func _gte(m *Mach) error {
	b, err := m.pop()
	if err != nil {
		return err
	}
	a, err := m.pop()
	if err != nil {
		return err
	}
	if a > b {
		return m.push(1)
	}
	return m.push(0)
}

func _not(m *Mach) error {
	addr, err := m.pAddr(0)
	if err != nil {
		return err
	}
	i, _, pg := m.pageFor(addr)
	if pg.d[i] == 0 {
		pg.d[i] = 1
	} else {
		pg.d[i] = 0
	}
	return nil
}

func _and(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	addr, err := m.pAddr(1)
	if err != nil {
		return err
	}
	i, _, pg := m.pageFor(addr)
	if pg.d[i] != 0 && val != 0 {
		pg.d[i] = 1
	} else {
		pg.d[i] = 0
	}
	return nil
}

func _or(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	addr, err := m.pAddr(1)
	if err != nil {
		return err
	}
	i, _, pg := m.pageFor(addr)
	if pg.d[i] != 0 || val != 0 {
		pg.d[i] = 1
	} else {
		pg.d[i] = 0
	}
	return nil
}

func _xor(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	addr, err := m.pAddr(1)
	if err != nil {
		return err
	}
	i, _, pg := m.pageFor(addr)
	if pg.d[i] != 0 && val == 0 {
		pg.d[i] = 1
	} else if pg.d[i] == 0 && val != 0 {
		pg.d[i] = 1
	} else {
		pg.d[i] = 0
	}
	return nil
}

func (arg _ltImm) run(m *Mach) error {
	a, err := m.pop()
	if err != nil {
		return err
	}
	if a < uint32(arg) {
		return m.push(1)
	}
	return m.push(0)
}

func (arg _lteImm) run(m *Mach) error {
	a, err := m.pop()
	if err != nil {
		return err
	}
	if a <= uint32(arg) {
		return m.push(1)
	}
	return m.push(0)
}

func (arg _eqImm) run(m *Mach) error {
	a, err := m.pop()
	if err != nil {
		return err
	}
	if a == uint32(arg) {
		return m.push(1)
	}
	return m.push(0)
}

func (arg _neqImm) run(m *Mach) error {
	a, err := m.pop()
	if err != nil {
		return err
	}
	if a != uint32(arg) {
		return m.push(1)
	}
	return m.push(0)
}

func (arg _gtImm) run(m *Mach) error {
	a, err := m.pop()
	if err != nil {
		return err
	}
	if a > uint32(arg) {
		return m.push(1)
	}
	return m.push(0)
}

func (arg _gteImm) run(m *Mach) error {
	a, err := m.pop()
	if err != nil {
		return err
	}
	if a >= uint32(arg) {
		return m.push(1)
	}
	return m.push(0)
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

func not(arg uint32, have bool) op {
	if have {
		return nil
	}
	return _not
}

func and(arg uint32, have bool) op {
	if have {
		return nil
	}
	return _and
}

func or(arg uint32, have bool) op {
	if have {
		return nil
	}
	return _or
}

func xor(arg uint32, have bool) op {
	if have {
		return nil
	}
	return _xor
}
