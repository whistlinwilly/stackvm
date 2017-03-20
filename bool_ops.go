package stackvm

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
	addr, err := m.pAddr(int(1))
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
	addr, err := m.pAddr(int(1))
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
	addr, err := m.pAddr(int(1))
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
