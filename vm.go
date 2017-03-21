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
	ip       int     // next op to decode
	pbp, psp int     // param stack
	cbp, csp int     // control stack
	pages    []*page // memory
}

type context interface {
	queue(*Mach) error
}

type page struct {
	r int32
	d [64]byte
}

func (pg *page) fetch(off int) byte {
	if pg == nil {
		return 0
	}
	return pg.d[off]
}

func (pg *page) store(off int, val byte) *page {
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

func (m *Mach) fetchBytes(addr int, bs []byte) (n int) {
	_, j, pg := m.pageFor(addr)
	for n < len(bs) {
		if j > 0x3f {
			addr += addr + 0x3f
			_, j, pg = m.pageFor(addr)
		}
		if pg == nil {
			left := len(pg.d) - j
			if len(bs)-n <= left {
				n++
				break
			}
			j += left
			n += left
			continue
		}
		bs[n] = pg.d[j]
		j++
		n++
	}
	return
}

func (m *Mach) decode(addr int) (end int, code byte, arg uint32, have bool, err error) {
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

func (m *Mach) jump(off int) error {
	ip := m.ip + off
	if ip >= m.pbp && ip <= m.cbp {
		return errSegfault
	}
	m.ip = ip
	return nil
}

func (m *Mach) fork(off int) error {
	if m.ctx == nil {
		return errNoConetxt
	}
	ip := m.ip + off
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

func (m *Mach) branch(off int) error {
	if m.ctx == nil {
		return errNoConetxt
	}
	ip := m.ip + off
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

func (m *Mach) call(ip int) error {
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

func (m *Mach) fetch(addr int) byte {
	_, j, pg := m.pageFor(addr)
	return pg.fetch(j)
}

func (m *Mach) store(addr int, val byte) {
	i, j, pg := m.pageFor(addr)
	npg := pg.store(j, val)
	if i >= len(m.pages) {
		pages := make([]*page, i+1)
		copy(pages, m.pages)
		m.pages = pages
	}
	if npg != pg {
		m.pages[i] = pg
	}
}

func (m *Mach) pageFor(addr int) (i, j int, pg *page) {
	i, j = addr>>6, addr&0x3f
	if i < len(m.pages) {
		pg = m.pages[i]
	}
	return
}

func (m *Mach) push(val byte) error {
	if m.psp < m.csp {
		m.store(m.psp, val)
		m.psp++
		return nil
	}
	return stackRangeError{"param", "over"}
}

func (m *Mach) pop() (byte, error) {
	if psp := m.psp - 1; psp >= m.pbp {
		m.psp = psp
		return m.fetch(psp), nil
	}
	return 0, stackRangeError{"param", "under"}
}

func (m *Mach) pAddr(off int) (int, error) {
	if addr := m.psp - off; addr >= m.pbp {
		return addr, nil
	}
	return 0, stackRangeError{"param", "under"}
}

func (m *Mach) cpush(val byte) error {
	if m.csp > m.psp {
		m.store(m.csp, val)
		m.csp--
		return nil
	}
	return stackRangeError{"control", "over"}
}

func (m *Mach) cpop() (byte, error) {
	if csp := m.csp + 1; csp <= m.cbp {
		m.csp = csp
		return m.fetch(csp), nil
	}
	return 0, stackRangeError{"control", "under"}
}

func (m *Mach) cAddr(off int) (int, error) {
	if addr := m.csp + off; addr <= m.cbp {
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
