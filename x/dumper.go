package xstackvm

import (
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

func (d *dumper) page(addr uint32, p [64]byte) error {
	if d.last < addr {
		d.f("........  .. .. .. .. .. .. .. ..  .. .. .. .. .. .. .. ..")
	}
	i := uint32(0)
	for i < 64 {
		j := i + 0x10
		if err := d.line(addr+i, p[i:j]); err != nil {
			return err
		}
		i = j
	}
	d.last += 64
	return nil
}

func (d dumper) line(addr uint32, l []byte) error {
	bs := d.formatBytes(addr, l)
	if ann := d.annotate(addr, l); ann != "" {
		d.f("%08x  %s  %s", addr, bs, ann)
	} else {
		d.f("%08x  %s", addr, bs)
	}
	return nil
}

func (d dumper) formatBytes(addr uint32, l []byte) string {
	return fmt.Sprintf(
		"%02x %02x %02x %02x %02x %02x %02x %02x  %02x %02x %02x %02x %02x %02x %02x %02x",
		l[0], l[1], l[2], l[3],
		l[4], l[5], l[6], l[7],
		l[8], l[9], l[10], l[11],
		l[12], l[13], l[14], l[15],
	)
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
	if ann := d.annotateStackBytes(addr, l, d.m.CBP(), d.m.CSP()); ann != "" {
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
	return fmt.Sprintf("%d", makeUint32s(l[lo-addr:hi-addr]))
}

func makeUint32s(p []byte) []uint32 {
	ns := make([]uint32, len(p)/4)
	for i := 0; i < len(p); i += 4 {
		ns[i/4] = makeUint32(p[i+0], p[i+1], p[i+2], p[i+3])
	}
	return ns
}

func makeUint32(a, b, c, d byte) uint32 {
	return uint32(a)<<24 | uint32(b)<<16 | uint32(c)<<8 | uint32(d)
}
