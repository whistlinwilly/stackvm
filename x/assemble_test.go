package xstackvm_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/jcorbin/stackvm/x"
)

type assemblerCases []assemblerCase

func (cs assemblerCases) run(t *testing.T) {
	for _, c := range cs {
		t.Run(c.name, c.run)
	}
}

type assemblerCase struct {
	name string
	in   []interface{}
	out  []byte
	err  error
}

func (c assemblerCase) run(t *testing.T) {
	assemblerTest{c, t}.run()
}

type assemblerTest struct {
	assemblerCase
	*testing.T
}

func (t assemblerTest) run() {
	prog, err := Assemble(t.in...)
	if t.err == nil {
		require.NoError(t, err, "unexpected error")
	} else {
		assert.EqualError(t, err, t.err.Error(), "expected error")
	}
	assert.Equal(t, t.out, prog, "expected machine code")
}

func TestAssemble(t *testing.T) {
	assemblerCases{
		{
			name: "basic",
			in: []interface{}{
				2, "push",
				3, "add",
				5, "eq",
			},
			out: []byte{
				0x82, 0x00,
				0x83, 0x11,
				0x85, 0x1a,
			},
		},

		{
			name: "small loop",
			in: []interface{}{
				10, "push",
				"loop:",
				1, "sub",
				0, "gt",
				":loop", "jnz",
				"halt",
			},
			out: []byte{
				0x8a, 0x00,
				0x81, 0x12,
				0x80, 0x1c,
				0x8f, 0xff, 0xff, 0xff, 0xf6, 0x29,
				0x7f,
			},
		},
	}.run(t)
}
