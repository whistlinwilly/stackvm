package stackvm_test

import (
	"fmt"
	"testing"

	. "github.com/jcorbin/stackvm/x"
)

func TestMach_basic_math(t *testing.T) {
	TestCases{
		{
			Name: "33addeq5 should fail",
			Err:  "HALT(1)",
			Prog: MustAssemble(
				0x40,
				3, "push", 3, "push", "add",
				5, "push", "eq",
				1, "hz", "halt",
			),
			Result: Result{
				Err: "HALT(1)",
			},
		},
		{
			Name: "23addeq5 should succeed",
			Prog: MustAssemble(
				0x40,
				2, "push", 3, "push", "add",
				5, "push", "eq",
				1, "hz", "halt",
			),
			Result: Result{},
		},
	}.Run(t)
}

func TestMach_collatz_sequence(t *testing.T) {
	tcs := make(TestCases, 0, 9)
	for n := 1; n < 10; n++ {
		vals := []uint32{uint32(n)}
		val := vals[0]
		for {
			switch {
			case val%2 == 0:
				val = val / 2
			default:
				val = 3*val + 1
			}
			vals = append(vals, val)
			if val <= 1 {
				break
			}
		}

		tcs = append(tcs, TestCase{
			Name: fmt.Sprintf("collatz(%d)", n),
			Prog: MustAssemble(
				0x40,

				n, "push", "dup", // v v :
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

				"c2p",         // v i :
				0x100, "push", // v i base :
				2, "p2c", // v : i base
				"halt",
			),
			Result: Result{
				Values: [][]uint32{vals},
			},
		})
	}

	tcs.Run(t)
}

func TestMach_collatz_explore(t *testing.T) {
	TestCase{
		Name: "gen collatz",
		Prog: MustAssemble(
			0x40,

			6, "push", // d :
			0x100, "push", // d i :
			0x100, "push", // d i b :
			3, "p2c", // : i d i
			1, "push", // v=1 : b i d

			"round:", // v : i d

			"dup", 1, "sub", 3, "mod", // v (v-1)%3 : i d
			":third", "fz", // v : i d
			"double:", 2, "mul", // v=2*v : i d
			":next", "jump", // ...
			"third:", 1, "sub", 3, "div", // v=(v-1)/3 : i d

			"next:",        // v : i d
			"dup", 1, "hz", // v : i d

			"dup",    // v v : i d
			2, "c2p", // v v d i :
			"dup", 4, "add", "p2c", // v v d i : i+=4
			"swap",  // v v i d : i
			"p2c",   // v v i : i d
			"store", // v : i d

			"c2p", 1, "sub", // v d-- : i
			"dup", "p2c", 0, "gt", // v d>0 : i d
			":round", "jnz", // v : i d

			"pop", "cpop", "halt", // : i
		),

		Result: Results{
			{Values: [][]uint32{{2, 4, 8, 16, 32, 64}}},
			{Values: [][]uint32{{2, 4, 8, 16, 5, 10}}},
			{Values: [][]uint32{{2, 4, 1, 2, 4, 8}}},
			{Values: [][]uint32{{2, 4, 1, 2, 4, 1}}},
		}.WithExpectedHaltCodes(1),
	}.Run(t)
}
