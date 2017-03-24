package stackvm_test

import (
	"testing"

	. "github.com/jcorbin/stackvm/x"
)

func TestMach(t *testing.T) {
	TestCases{
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
	}.Run(t)
}
