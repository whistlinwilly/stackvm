package stackvm

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func buildFilledPages(n int) []*page {
	r := make([]*page, n)
	v := 0
	for i := 0; i < n; i++ {
		pg := &page{r: 1}
		r[i] = pg
		for i := 0; i < len(pg.d); i++ {
			pg.d[i] = byte(1 + v%255)
			v++
		}
	}
	return r
}

func TestMach_fetchBytes(t *testing.T) {
	m := Mach{pages: buildFilledPages(3)}
	for stride := 1; stride <= 2*_pageSize; stride++ {
		t.Run(fmt.Sprintf("stride %d", stride), func(t *testing.T) {
			expected := make([]byte, stride)
			v := 0
			for i := 0; i < stride; i++ {
				expected[i] = byte(1 + v%255)
				v++
			}
			for addr, n := 0, len(m.pages)*_pageSize; addr < n; addr += stride {
				buf := make([]byte, stride)
				actual := buf[:m.fetchBytes(uint32(addr), buf)]
				assert.Equal(t, []byte(expected), actual, "fetchBytes(%04x, %d)", addr, len(buf))
				for i := 0; i < len(expected); i++ {
					if v < n {
						expected[i] = byte(1 + v%255)
						v++
					} else {
						expected[i] = 0
					}
				}
			}
		})
	}
}
