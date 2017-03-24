package stackvm

// TODO: control-stack-addressed variants?

type _fetchImm uint32
type _storeImm uint32

func (arg _fetchImm) run(m *Mach) error {
	addr := uint32(arg)
	val, err := m.fetch(addr)
	if err != nil {
		return err
	}
	return m.push(val)
}

func (arg _storeImm) run(m *Mach) error {
	addr := uint32(arg)
	val, err := m.pop()
	if err != nil {
		return err
	}
	return m.store(addr, val)
}

func _fetch(m *Mach) error {
	addr, err := m.pop()
	if err != nil {
		return err
	}
	val, err := m.fetch(addr)
	if err != nil {
		return err
	}
	return m.push(val)
}

func _store(m *Mach) error {
	addr, err := m.pop()
	if err != nil {
		return err
	}
	val, err := m.pop()
	if err != nil {
		return err
	}
	return m.store(addr, val)
}

func fetch(arg uint32, have bool) op {
	if have {
		return _fetchImm(arg).run
	}
	return _fetch
}

func store(arg uint32, have bool) op {
	if have {
		return _storeImm(arg).run
	}
	return _store
}
