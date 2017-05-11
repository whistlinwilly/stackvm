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
	m.pa = bool2uint32(m.pa < b)
	return nil
}

func _lte(m *Mach) error {
	b, err := m.pop()
	if err != nil {
		return err
	}
	m.pa = bool2uint32(m.pa <= b)
	return nil
}

func _eq(m *Mach) error {
	b, err := m.pop()
	if err != nil {
		return err
	}
	m.pa = bool2uint32(m.pa == b)
	return nil
}

func _neq(m *Mach) error {
	b, err := m.pop()
	if err != nil {
		return err
	}
	m.pa = bool2uint32(m.pa != b)
	return nil
}

func _gt(m *Mach) error {
	b, err := m.pop()
	if err != nil {
		return err
	}
	m.pa = bool2uint32(m.pa > b)
	return nil
}

func _gte(m *Mach) error {
	b, err := m.pop()
	if err != nil {
		return err
	}
	m.pa = bool2uint32(m.pa >= b)
	return nil
}

func _not(m *Mach) error {
	m.pa = bool2uint32(m.pa == 0)
	return nil
}

func _and(m *Mach) error {
	b, err := m.pop()
	if err != nil {
		return err
	}
	m.pa = bool2uint32((m.pa != 0) && (b != 0))
	return nil
}

func _or(m *Mach) error {
	b, err := m.pop()
	if err != nil {
		return err
	}
	m.pa = bool2uint32((m.pa != 0) || (b != 0))
	return nil
}

func (arg _ltImm) run(m *Mach) error {
	m.pa = bool2uint32(m.pa < uint32(arg))
	return nil
}

func (arg _lteImm) run(m *Mach) error {
	m.pa = bool2uint32(m.pa <= uint32(arg))
	return nil
}

func (arg _eqImm) run(m *Mach) error {
	m.pa = bool2uint32(m.pa == uint32(arg))
	return nil
}

func (arg _neqImm) run(m *Mach) error {
	m.pa = bool2uint32(m.pa != uint32(arg))
	return nil
}

func (arg _gtImm) run(m *Mach) error {
	m.pa = bool2uint32(m.pa > uint32(arg))
	return nil
}

func (arg _gteImm) run(m *Mach) error {
	m.pa = bool2uint32(m.pa >= uint32(arg))
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

func bool2uint32(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}
