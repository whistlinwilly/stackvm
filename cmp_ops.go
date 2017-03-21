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
