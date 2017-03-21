package stackvm

import (
	"errors"
	"fmt"
	"sync/atomic"
)

var (
	errVarIntTooBig  = errors.New("varint argument too big")
	errInvalidIP     = errors.New("invalid IP")
	errSegfault      = errors.New("segfault")
	errNoConetxt     = errors.New("no context, cannot copy")
	errUnimplemented = errors.New("unipmlemented")
)

// Mach is a stack machine
type Mach struct {
	ctx      context // execution context
	err      error   // non-nil after termination
	ip       uint32  // next op to decode
	pbp, psp uint32  // param stack
	cbp, csp uint32  // control stack
	pages    []*page // memory
}

type context interface {
	queue(*Mach) error
}

type page struct {
	r int32
	d [64]byte
}

func (pg *page) fetch(off uint32) byte {
	if pg == nil {
		return 0
	}
	return pg.d[off]
}

func (pg *page) store(off uint32, val byte) *page {
	if pg == nil {
		pg = &page{r: 1}
	} else if r := atomic.LoadInt32(&pg.r); r > 1 {
		newPage := &page{r: 1, d: pg.d}
		pg.decref()
		pg = newPage
	}
	pg.d[off] = val
	return pg
}

func (pg *page) decref() {
	if pg != nil {
		atomic.AddInt32(&pg.r, 1)
	}
}

func (pg *page) incref() {
	if pg != nil {
		atomic.AddInt32(&pg.r, 1)
	}
}

func (m *Mach) run() {
	for m.err == nil {
		m.step()
	}
}

func (m *Mach) step() {
	ip, code, arg, have, err := m.decode(m.ip)
	if err != nil {
		m.err = err
		return
	}
	op, err := makeOp(code, arg, have)
	if err != nil {
		m.err = err
		return
	}
	m.ip = ip
	if err := op(m); err != nil {
		m.err = err
		return
	}
}

func (m *Mach) decode(addr uint32) (end uint32, code byte, arg uint32, have bool, err error) {
	var bs [5]byte
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
	if n < 5 {
		err = errInvalidIP
	} else {
		err = errVarIntTooBig
	}
	return
}

func (m *Mach) jump(off int32) error {
	return m.jumpTo(uint32(int32(m.ip) + off))
}

func (m *Mach) jumpTo(ip uint32) error {
	if ip >= m.pbp && ip <= m.cbp {
		return errSegfault
	}
	m.ip = ip
	return nil
}

func (m *Mach) fork(off int32) error {
	if m.ctx == nil {
		return errNoConetxt
	}
	ip := uint32(int32(m.ip) + off)
	if ip >= m.pbp && ip <= m.cbp {
		return errSegfault
	}
	n := *m
	n.pages = n.pages[:len(n.pages):len(n.pages)]
	for _, pg := range n.pages {
		pg.incref()
	}
	m.ip = ip
	return m.ctx.queue(&n)
}

func (m *Mach) branch(off int32) error {
	if m.ctx == nil {
		return errNoConetxt
	}
	ip := uint32(int32(m.ip) + off)
	if ip >= m.pbp && ip <= m.cbp {
		return errSegfault
	}
	n := *m
	n.pages = n.pages[:len(n.pages):len(n.pages)]
	for _, pg := range n.pages {
		pg.incref()
	}
	n.ip = ip
	return m.ctx.queue(&n)
}

func (m *Mach) call(ip uint32) error {
	return errUnimplemented // FIXME ip int vs byte memory
	// if ip >= m.pbp && ip <= m.cbp {
	// 	return errSegfault
	// }
	// if err := m.cpush(m.ip); err != nil {
	// 	return err
	// }
	// m.ip = ip
	// return nil
}

func (m *Mach) ret() error {
	return errUnimplemented // FIXME ip int vs byte memory
	// ip, err := m.cpop()
	// if err != nil {
	// 	return err
	// }
	// m.ip = ip
	// return nil
}

func (m *Mach) fetchBytes(addr uint32, bs []byte) (n int) {
	_, j, pg := m.pageFor(addr)
	for n < len(bs) {
		if j > 0x3f {
			addr += addr + 0x3f
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
	for n := 0; n < len(bs); n++ {
		if j > 0x3f {
			addr += addr + 0x3f
			i, j, pg = m.pageFor(addr)
		}
		npg := pg.store(j, bs[n])
		if int(i) >= len(m.pages) {
			pages := make([]*page, i+1)
			copy(pages, m.pages)
			m.pages = pages
		}
		if npg != pg {
			pg, m.pages[i] = npg, npg
		}
	}
}

func (m *Mach) fetch(addr uint32) (uint32, error) {
	_, j, pg := m.pageFor(addr)
	return pg.fetch(j)
}

func (m *Mach) store(addr, val uint32) error {
	i, j, pg := m.pageFor(addr)
	if npg, err := pg.store(j, val); err != nil {
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

func (m *Mach) pageFor(addr uint32) (i, j uint32, pg *page) {
	i, j = addr>>6, addr&0x3f
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
	if psp := m.psp - 4; psp >= m.pbp {
		m.psp = psp
		return m.fetch(psp)
	}
	return 0, stackRangeError{"param", "under"}
}

func (m *Mach) pAddr(i int32) (uint32, error) {
	if addr := uint32(int32(m.psp) - (i+1)*4); addr >= m.pbp {
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
	if csp := m.csp + 4; csp <= m.cbp {
		m.csp = csp
		return m.fetch(csp)
	}
	return 0, stackRangeError{"control", "under"}
}

func (m *Mach) cAddr(i int32) (uint32, error) {
	if addr := uint32(int32(m.csp) - (i+1)*4); addr >= m.cbp {
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
