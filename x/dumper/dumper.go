package dumper

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/jcorbin/stackvm"
)

type dumper struct {
	m    *stackvm.Mach
	f    func(string, ...interface{})
	last uint32
}

// Dump dumps the machine's memory to a log formating function.
func Dump(m *stackvm.Mach, f func(string, ...interface{})) error {
	d := dumper{
		m: m,
		f: f,
	}
	return m.EachPage(d.page)
}

func (d *dumper) page(addr uint32, p *[64]byte) error {
	if d.last < addr {
		d.f("........  .. .. .. .. .. .. .. ..  .. .. .. .. .. .. .. ..")
	}
	i := uint32(0)
	for i < 64 {
		j := i + 0x10
		d.line(addr+i, p[i:j])
		i = j
	}
	d.last = addr + 64
	return nil
}

func (d dumper) line(addr uint32, l []byte) {
	bs := fmt.Sprintf(
		"%02x %02x %02x %02x %02x %02x %02x %02x  %02x %02x %02x %02x %02x %02x %02x %02x",
		l[0], l[1], l[2], l[3],
		l[4], l[5], l[6], l[7],
		l[8], l[9], l[10], l[11],
		l[12], l[13], l[14], l[15],
	)
	if ann := d.annotate(addr, l); ann != "" {
		d.f("%08x  %s  %s", addr, bs, ann)
	} else {
		d.f("%08x  %s", addr, bs)
	}
}

func (d dumper) annotate(addr uint32, l []byte) string {
	var (
		parts [2]string
		i     int
	)
	if ann := d.annotateStackBytes(addr, l, d.m.PBP(), d.m.PSP()); ann != "" {
		parts[i] = ann
		i++
	}
	if ann := d.annotateStackBytes(addr, l, d.m.CSP()+4, d.m.CBP()+4); ann != "" {
		parts[i] = ann
		i++
	}
	if i > 0 {
		return strings.Join(parts[:i], " ")
	}

	return ""
}

func (d dumper) annotateStackBytes(addr uint32, l []byte, bp, sp uint32) string {
	lo, hi := addr, addr+uint32(len(l))
	if bp > lo {
		lo = bp
	}
	if sp < hi {
		hi = sp
	}
	if lo >= hi {
		return ""
	}
	p := l[lo-addr : hi-addr]
	ns := make([]uint32, len(p)/4)
	for i, j := 0, 0; j < len(p); i, j = i+1, j+4 {
		ns[i] = binary.LittleEndian.Uint32(p[j:]) // XXX architecture dependent
	}
	return fmt.Sprintf("%d", ns)
}
