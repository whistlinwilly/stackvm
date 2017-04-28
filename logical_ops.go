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
	ap, err := m.pRef(1)
	if err != nil {
		return err
	}
	return setBool(ap, *ap < b)
}

func _lte(m *Mach) error {
	b, err := m.pop()
	if err != nil {
		return err
	}
	ap, err := m.pRef(1)
	if err != nil {
		return err
	}
	return setBool(ap, *ap <= b)
}

func _eq(m *Mach) error {
	b, err := m.pop()
	if err != nil {
		return err
	}
	ap, err := m.pRef(1)
	if err != nil {
		return err
	}
	return setBool(ap, *ap == b)
}

func _neq(m *Mach) error {
	b, err := m.pop()
	if err != nil {
		return err
	}
	ap, err := m.pRef(1)
	if err != nil {
		return err
	}
	return setBool(ap, *ap != b)
}

func _gt(m *Mach) error {
	b, err := m.pop()
	if err != nil {
		return err
	}
	ap, err := m.pRef(1)
	if err != nil {
		return err
	}
	return setBool(ap, *ap > b)
}

func _gte(m *Mach) error {
	b, err := m.pop()
	if err != nil {
		return err
	}
	ap, err := m.pRef(1)
	if err != nil {
		return err
	}
	return setBool(ap, *ap >= b)
}

func _not(m *Mach) error {
	p, err := m.pRef(1)
	if err != nil {
		return err
	}
	return setBool(p, *p == 0)
}

func _and(m *Mach) error {
	b, err := m.pop()
	if err != nil {
		return err
	}
	ap, err := m.pRef(1)
	if err != nil {
		return err
	}
	return setBool(ap, (*ap != 0) && (b != 0))
}

func _or(m *Mach) error {
	b, err := m.pop()
	if err != nil {
		return err
	}
	ap, err := m.pRef(1)
	if err != nil {
		return err
	}
	return setBool(ap, (*ap != 0) || (b != 0))
}

func (arg _ltImm) run(m *Mach) error {
	ap, err := m.pRef(1)
	if err != nil {
		return err
	}
	return setBool(ap, *ap < uint32(arg))
}

func (arg _lteImm) run(m *Mach) error {
	ap, err := m.pRef(1)
	if err != nil {
		return err
	}
	return setBool(ap, *ap <= uint32(arg))
}

func (arg _eqImm) run(m *Mach) error {
	ap, err := m.pRef(1)
	if err != nil {
		return err
	}
	return setBool(ap, *ap == uint32(arg))
}

func (arg _neqImm) run(m *Mach) error {
	ap, err := m.pRef(1)
	if err != nil {
		return err
	}
	return setBool(ap, *ap != uint32(arg))
}

func (arg _gtImm) run(m *Mach) error {
	ap, err := m.pRef(1)
	if err != nil {
		return err
	}
	return setBool(ap, *ap > uint32(arg))
}

func (arg _gteImm) run(m *Mach) error {
	ap, err := m.pRef(1)
	if err != nil {
		return err
	}
	return setBool(ap, *ap >= uint32(arg))
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

func setBool(p *uint32, b bool) error {
	if b {
		*p = 1
	} else {
		*p = 0
	}
	return nil
}
