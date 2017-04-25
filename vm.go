package stackvm

import (
	"errors"
	"fmt"
	"sync/atomic"
)

const (
	_pageSize        = 0x40
	_pageMask        = _pageSize - 1
	_machVersionCode = 0x00
)

var (
	errVarIntTooBig = errors.New("varint argument too big")
	errInvalidIP    = errors.New("invalid IP")
	errSegfault     = errors.New("segfault")
	errNoQueue      = errors.New("no queue, cannot copy")
	errAlignment    = errors.New("unaligned memory access")
)

type alignmentError struct {
	op   string
	addr uint32
}

func (ae alignmentError) Error() string {
	return fmt.Sprintf("unaligned memory %s @0x%04x", ae.op, ae.addr)
}

// Mach is a stack machine.
type Mach struct {
	ctx      context // execution context
	err      error   // non-nil after termination
	ip       uint32  // next op to decode
	pbp, psp uint32  // param stack
	cbp, csp uint32  // control stack
	// TODO track code segment and data segment
	pages []*page // memory
}

type page struct {
	r int32
	d [_pageSize]byte
}

func (pg *page) fetchByte(off uint32) byte {
	if pg == nil {
		return 0
	}
	return pg.d[off]
}

func (pg *page) fetch(off uint32) (uint32, error) {
	if off%4 != 0 {
		return 0, errAlignment
	}
	if pg == nil {
		return 0, nil
	}
	val := uint32(pg.d[off+0])<<24 | uint32(pg.d[off+1])<<16 | uint32(pg.d[off+2])<<8 | uint32(pg.d[off+3])
	return val, nil
}

func (pg *page) own() *page {
	if pg == nil {
		return &page{r: 1}
	}
	if atomic.LoadInt32(&pg.r) == 1 {
		return pg
	}
	newPage := &page{r: 1, d: pg.d}
	atomic.AddInt32(&pg.r, 1)
	return newPage
}

func (pg *page) storeByte(off uint32, val byte) *page {
	pg = pg.own()
	pg.d[off] = val
	return pg
}

func (pg *page) store(off uint32, val uint32) (*page, error) {
	if off%4 != 0 {
		return nil, errAlignment
	}
	pg = pg.own()
	pg.d[off] = byte((val >> 24) & 0xff)
	pg.d[off+1] = byte((val >> 16) & 0xff)
	pg.d[off+2] = byte((val >> 8) & 0xff)
	pg.d[off+3] = byte(val & 0xff)
	return pg, nil
}

func (m *Mach) halted() (uint32, bool) {
	arg, ok := m.err.(_halt)
	return uint32(arg), ok
}

func (m *Mach) run() (*Mach, error) {

repeat:
	// live
	for m.err == nil {
		m.step()
	}

	// win or die
	err := m.ctx.Handle(m)
	if err == nil {
		if n := m.ctx.next(); n != nil {
			m = n
			// die
			goto repeat
		}
	}

	// win?
	return m, err
}

func (m *Mach) step() {
	op, err := m.decode()
	if err == nil {
		err = op(m)
	}
	m.err = err
}

func (m *Mach) decode() (op, error) {
	ip, code, arg, have, err := m.read(m.ip)
	if err != nil {
		return nil, err
	}
	op, err := makeOp(code, arg, have)
	if err != nil {
		return nil, err
	}
	m.ip = ip
	return op, nil
}

func (m *Mach) read(addr uint32) (end uint32, code byte, arg uint32, have bool, err error) {
	var bs [6]byte
	end = addr
	n := m.fetchBytes(addr, bs[:])
	for k := 0; k < n; k++ {
		val := bs[k]
		end++
		if val&0x80 == 0 {
			code = val
			have = k > 0
			return
		}
		if k == len(bs)-1 {
			break
		}
		arg = arg<<7 | uint32(val&0x7f)
	}
	if n < len(bs) {
		err = errInvalidIP
	} else {
		err = errVarIntTooBig
	}
	return
}

func (m *Mach) jump(off int32) error {
	return m.jumpTo(uint32(int32(m.ip) + off))
}

func (m *Mach) cjump() error {
	ip, err := m.cpop()
	if err != nil {
		return err
	}
	return m.jumpTo(ip)
}

func (m *Mach) jumpTo(ip uint32) error {
	if ip >= m.pbp && ip <= m.cbp {
		return errSegfault
	}
	m.ip = ip
	return nil
}

func (m *Mach) copy() (*Mach, error) {
	n := *m
	n.pages = make([]*page, len(n.pages))
	for i, pg := range m.pages {
		if pg != nil {
			n.pages[i] = pg
			atomic.AddInt32(&pg.r, 1)
		}
	}
	return &n, nil
}

func (m *Mach) fork(off int32) error {
	ip := uint32(int32(m.ip) + off)
	if ip >= m.pbp && ip <= m.cbp {
		return errSegfault
	}
	n, err := m.copy()
	if err != nil {
		return err
	}
	n.ip = ip
	return m.ctx.queue(n)
}

func (m *Mach) cfork() error {
	n, err := m.copy()
	if err != nil {
		return err
	}
	ip, err := n.cpop()
	if err != nil {
		return err
	}
	if err := n.jumpTo(ip); err != nil {
		return err
	}
	return m.ctx.queue(n)
}

func (m *Mach) branch(off int32) error {
	ip := uint32(int32(m.ip) + off)
	if ip >= m.pbp && ip <= m.cbp {
		return errSegfault
	}
	n, err := m.copy()
	if err != nil {
		return err
	}
	m.ip = ip
	return m.ctx.queue(n)
}

func (m *Mach) cbranch() error {
	n, err := m.copy()
	if err != nil {
		return err
	}
	ip, err := m.cpop()
	if err != nil {
		return err
	}
	if err := m.ctx.queue(n); err != nil {
		return err
	}
	return m.jumpTo(ip)
}

func (m *Mach) loop() error {
	addr, err := m.cAddr(0)
	if err != nil {
		return err
	}
	ip, err := m.fetch(addr)
	if err != nil {
		return err
	}
	return m.jumpTo(ip)
}

func (m *Mach) call(ip uint32) error {
	if ip >= m.pbp && ip <= m.cbp {
		return errSegfault
	}
	if err := m.cpush(m.ip); err != nil {
		return err
	}
	return m.jumpTo(ip)
}

func (m *Mach) ret() error {
	ip, err := m.cpop()
	if err != nil {
		return err
	}
	return m.jumpTo(ip)
}

func (m *Mach) fetchPS() ([]uint32, error) {
	psp := m.psp
	if psp > m.csp {
		psp = m.csp
	}
	if psp > m.cbp {
		psp = m.cbp
	}
	return m.fetchMany(m.pbp, psp)
}

func (m *Mach) fetchCS() ([]uint32, error) {
	csp := m.csp
	if csp > m.csp {
		csp = m.csp
	}
	if csp > m.cbp {
		csp = m.cbp
	}
	return m.fetchMany(m.cbp, csp)
}

func (m *Mach) fetchMany(from, to uint32) ([]uint32, error) {
	if from == to {
		return nil, nil
	}

	if from < to {
		ns := make([]uint32, 0, (to-from)/4)
		for ; from < to; from += 4 {
			val, err := m.fetch(from)
			if err != nil {
				return nil, err
			}
			ns = append(ns, val)
		}
		return ns, nil
	}

	// to < from
	ns := make([]uint32, 0, (from-to)/4)
	for ; from > to; from -= 4 {
		val, err := m.fetch(from)
		if err != nil {
			return nil, err
		}
		ns = append(ns, val)
	}
	return ns, nil
}

func (m *Mach) fetchBytes(addr uint32, bs []byte) (n int) {
	_, j, pg := m.pageFor(addr)
	for n < len(bs) {
		if j > _pageMask {
			addr += _pageSize
			_, j, pg = m.pageFor(addr)
		}
		if pg == nil {
			left := len(pg.d) - int(j)
			if len(bs)-n <= left {
				n++
				break
			}
			j += uint32(left)
			n += left
			continue
		}
		bs[n] = pg.d[j]
		j++
		n++
	}
	return
}

func (m *Mach) storeBytes(addr uint32, bs []byte) {
	i, j, pg := m.pageFor(addr)
	// TODO: pg.storeBytes(addr, bs) int
	for n := 0; n < len(bs); n++ {
		if j > _pageMask {
			addr += _pageSize
			i, j, pg = m.pageFor(addr)
		}
		npg := pg.storeByte(j, bs[n])
		if int(i) >= len(m.pages) {
			pages := make([]*page, i+1)
			copy(pages, m.pages)
			m.pages = pages
		}
		if npg != pg {
			pg, m.pages[i] = npg, npg
		}
		j++
	}
}

func (m *Mach) fetch(addr uint32) (uint32, error) {
	_, j, pg := m.pageFor(addr)
	val, err := pg.fetch(j)
	if err == errAlignment {
		err = alignmentError{"fetch", addr}
	}
	return val, err
}

func (m *Mach) store(addr, val uint32) error {
	i, j, pg := m.pageFor(addr)
	if npg, err := pg.store(j, val); err != nil {
		if err == errAlignment {
			err = alignmentError{"store", addr}
		}
		return err
	} else if npg != pg {
		if int(i) >= len(m.pages) {
			pages := make([]*page, i+1)
			copy(pages, m.pages)
			m.pages = pages
		}
		pg, m.pages[i] = npg, npg

	}
	return nil
}

func (m *Mach) move(src, dst uint32) error {
	val, err := m.fetch(src)
	if err != nil {
		return err
	}
	return m.store(dst, val)
}

func (m *Mach) pageFor(addr uint32) (i, j uint32, pg *page) {
	i, j = addr>>6, addr&_pageMask
	if int(i) < len(m.pages) {
		pg = m.pages[i]
	}
	return
}

func (m *Mach) push(val uint32) error {
	if psp := m.psp + 4; psp <= m.csp {
		if err := m.store(m.psp, val); err != nil {
			return err
		}
		m.psp = psp
		return nil
	}
	return stackRangeError{"param", "over"}
}

func (m *Mach) pop() (uint32, error) {
	if m.psp <= m.pbp {
		return 0, stackRangeError{"param", "under"}
	}
	psp := m.psp - 4
	m.psp = psp
	return m.fetch(psp)
}

func (m *Mach) drop() error {
	if psp := m.psp - 4; psp >= m.pbp {
		m.psp = psp
		return nil
	}
	return stackRangeError{"param", "under"}
}

func (m *Mach) pAddr(i int32) (uint32, error) {
	if addr := uint32(int32(m.psp) - i*4); addr >= m.pbp && addr <= m.csp {
		return addr, nil
	}
	return 0, stackRangeError{"param", "under"}
}

func (m *Mach) cpush(val uint32) error {
	if csp := m.csp - 4; csp >= m.psp {
		if err := m.store(m.csp, val); err != nil {
			return err
		}
		m.csp = csp
		return nil
	}
	return stackRangeError{"control", "over"}
}

func (m *Mach) cpop() (uint32, error) {
	if m.csp >= m.cbp {
		return 0, stackRangeError{"control", "under"}
	}
	csp := m.csp + 4
	m.csp = csp
	return m.fetch(csp)
}

func (m *Mach) cdrop() error {
	if csp := m.csp + 4; csp <= m.cbp {
		m.csp = csp
		return nil
	}
	return stackRangeError{"control", "under"}
}

func (m *Mach) cAddr(i int32) (uint32, error) {
	if addr := uint32(int32(m.csp) + i*4); addr <= m.cbp && addr >= m.psp {
		return addr, nil
	}
	return 0, stackRangeError{"code", "under"}
}

type stackRangeError struct {
	name string
	kind string
}

func (sre stackRangeError) Error() string {
	return fmt.Sprintf("%s stack %sflow", sre.name, sre.kind)
}
