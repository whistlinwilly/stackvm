package stackvm_test

import (
	"testing"

	. "github.com/jcorbin/stackvm/x"
)

const (
	push = 0x00
	add  = 0x0a
	eq   = 0x13
	halt = 0x7f
)

func imm(v int) byte {
	if v > 0x7f {
		panic("nope imm")
	}
	return 0x80 | byte(v)
}

func TestMach(t *testing.T) {
	TestCases{
		{
			Name:      "23add5eq",
			StackSize: 64,
			Prog: []byte{
				imm(2), push, imm(3), push, add,
				imm(5), push, eq,
				halt,
			},
			Result: Result{
				PS: []uint32{1},
				// CS: []uint32{},
				// Mem []ResultMem
			},
		},
	}.Run(t)
}
