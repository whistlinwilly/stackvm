package stackvm_test

import (
	"fmt"
	"testing"

	. "github.com/jcorbin/stackvm/x"
)

func TestMach_basic_math(t *testing.T) {
	TestCases{
		{
			Name:      "33addeq5 should fail",
			StackSize: 64,
			Err:       "HALT(1)",
			Prog: MustAssemble(
				3, "push", 3, "push", "add",
				5, "push", "eq",
				":fail", "jz",
				"halt",
				"fail:", 1, "halt",
			),
			Result: Result{
				Err: "HALT(1)",
			},
		},
		{
			Name:      "23addeq5 should succeed",
			StackSize: 64,
			Prog: MustAssemble(
				2, "push", 3, "push", "add",
				5, "push", "eq",
				":fail", "jz",
				"halt",
				"fail:", 1, "halt",
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
			Name:      fmt.Sprintf("collatz(%d)", n),
			StackSize: 64,
			Prog: MustAssemble(
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
