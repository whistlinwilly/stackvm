package stackvm_test

import (
	"testing"

	. "github.com/jcorbin/stackvm/x"
)

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
				Mem: []ResultMem{
					{
						Addr: 0x100,
						Data: []byte{
							0x00, 0x00, 0x00, 0x09, // 9
							0x00, 0x00, 0x00, 0x1c, // 28
							0x00, 0x00, 0x00, 0x0e, // 14
							0x00, 0x00, 0x00, 0x07, // 7
							0x00, 0x00, 0x00, 0x16, // 22
							0x00, 0x00, 0x00, 0x0b, // 11
							0x00, 0x00, 0x00, 0x22, // 34
							0x00, 0x00, 0x00, 0x11, // 17,
							0x00, 0x00, 0x00, 0x34, // 52,
							0x00, 0x00, 0x00, 0x1a, // 26,
							0x00, 0x00, 0x00, 0x0d, // 13,
							0x00, 0x00, 0x00, 0x28, // 40,
							0x00, 0x00, 0x00, 0x14, // 20,
							0x00, 0x00, 0x00, 0x0a, // 10,
							0x00, 0x00, 0x00, 0x05, // 5,
							0x00, 0x00, 0x00, 0x10, // 16,
							0x00, 0x00, 0x00, 0x08, // 8,
							0x00, 0x00, 0x00, 0x04, // 4,
							0x00, 0x00, 0x00, 0x02, // 2,
							0x00, 0x00, 0x00, 0x01, // 1,
						},
					},
				},
			},
		},
	}.Run(t)
}
