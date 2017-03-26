package stackvm_test

import (
	"testing"

	. "github.com/jcorbin/stackvm/x"
)

func resNs(addr uint32, ns ...uint32) ResultMem {
	d := make([]byte, 0, len(ns)*4)
	for _, n := range ns {
		d = append(d,
			byte((n>>24)&0xff),
			byte((n>>16)&0xff),
			byte((n>>8)&0xff),
			byte((n)&0xff))
	}
	return ResultMem{
		Addr: addr,
		Data: d,
	}
}

func TestMach(t *testing.T) {
	TestCases{
		// basic math
		{
			Name:      "23add5eq",
			StackSize: 64,
			Prog: MustAssemble(
				2, "push", 3, "push", "add",
				5, "push", "eq",
				"halt",
			),
			Result: Result{
				PS: []uint32{1},
			},
		},

		// tracing collatz
		{
			Name:      "collatz_9",
			StackSize: 64,
			Prog: MustAssemble(
				9, "push", "dup", // v v :
				0x100, "push", // v v i :
				"dup", 4, "add", "p2c", // v v i : i=i+4
				"store", // v : i

				"loop:",         // v : i
				"dup", 2, "mod", // v v%2 : ...

				":odd", "jnz",

				"even:",
				2, "div", // v/2 : ...
				":next", "jump",

				"odd:",
				3, "mul", 1, "add", // 3*v+1 : ...

				"next:",
				"dup",    // v v : i
				"c2p",    // v v i :
				"dup",    // v v i i :
				4, "add", // v v i i+4 :
				"p2c",   // v v i : i=i+4
				"store", // v : i
				"dup",   // v v : i
				1, "eq", // v v==1 : i
				":loop", "jz", // v : i

				"halt",
			),
			Result: Result{
				PS: []uint32{1},
				CS: []uint32{0x100 + 20*4},
				Mem: []ResultMem{resNs(0x100,
					9,
					28, 14, 7,
					22, 11,
					34, 17,
					52, 26, 13,
					40, 20, 10, 5,
					16, 8, 4, 2, 1)},
			},
		},
	}.Run(t)
}
