package stackvm

type _addImm uint32
type _subImm uint32
type _mulImm uint32
type _divImm uint32
type _modImm uint32
type _divmodImm uint32

func _neg(m *Mach) error {
	a, err := m.pop()
	if err != nil {
		return err
	}
	return m.push(-a)
}

func _add(m *Mach) error {
	b, err := m.pop()
	if err != nil {
		return err
	}
	a, err := m.pop()
	if err != nil {
		return err
	}
	return m.push(a + b)
}

func _sub(m *Mach) error {
	b, err := m.pop()
	if err != nil {
		return err
	}
	a, err := m.pop()
	if err != nil {
		return err
	}
	return m.push(a - b)
}

func _mul(m *Mach) error {
	b, err := m.pop()
	if err != nil {
		return err
	}
	a, err := m.pop()
	if err != nil {
		return err
	}
	return m.push(a * b)
}

func _div(m *Mach) error {
	b, err := m.pop()
	if err != nil {
		return err
	}
	a, err := m.pop()
	if err != nil {
		return err
	}
	return m.push(a / b)
}

func _mod(m *Mach) error {
	b, err := m.pop()
	if err != nil {
		return err
	}
	a, err := m.pop()
	if err != nil {
		return err
	}
	return m.push(uint32(rem(int32(a), int32(b))))
}

func _divmod(m *Mach) error {
	b, err := m.pop()
	if err != nil {
		return err
	}
	a, err := m.pop()
	if err != nil {
		return err
	}
	if err := m.push(a / b); err != nil {
		return err
	}
	return m.push(uint32(rem(int32(a), int32(b))))
}

func (arg _addImm) run(m *Mach) error {
	a, err := m.pop()
	if err != nil {
		return err
	}
	return m.push(a + uint32(arg))
}

func (arg _subImm) run(m *Mach) error {
	a, err := m.pop()
	if err != nil {
		return err
	}
	return m.push(a - uint32(arg))
}

func (arg _mulImm) run(m *Mach) error {
	a, err := m.pop()
	if err != nil {
		return err
	}
	return m.push(a * uint32(arg))
}

func (arg _divImm) run(m *Mach) error {
	a, err := m.pop()
	if err != nil {
		return err
	}
	return m.push(a / uint32(arg))
}

func (arg _modImm) run(m *Mach) error {
	a, err := m.pop()
	if err != nil {
		return err
	}
	return m.push(uint32(rem(int32(a), int32(arg))))
}

func (arg _divmodImm) run(m *Mach) error {
	a, err := m.pop()
	if err != nil {
		return err
	}
	if err := m.push(a / uint32(arg)); err != nil {
		return err
	}
	return m.push(uint32(rem(int32(a), int32(arg))))
}

func neg(arg uint32, have bool) op {
	if have {
		return nil
	}
	return _neg
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

func rem(a, b int32) int32 {
	x := a % b
	if x < 0 {
		x += b
	}
	return x
}
