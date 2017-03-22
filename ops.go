package stackvm

import "fmt"

type op func(*Mach) error

type opDecoder func(arg uint32, have bool) op

var opCodes = [128]opDecoder{
	push, pop, dup, swap, nil, nil, nil, nil,
	neg, add, sub, mul, div, mod, divmod, nil,
	lt, lte, eq, neq, gt, gte, nil, nil,
	not, and, or, xor, nil, nil, nil, nil,
	jump, jnz, jz, nil, nil, nil, nil, nil,
	mark, call, ret, nil, nil, nil, nil, nil,
	cpop, p2c, c2p, nil, nil, nil, nil, nil,
	loop, lnz, lz, nil, nil, nil, nil, nil,
	fork, fnz, fz, nil, nil, nil, nil, nil,
	branch, bnz, bz, nil, nil, nil, nil, nil,
	nil, nil, nil, nil, nil, nil, nil, nil,
	nil, nil, nil, nil, nil, nil, nil, nil,
	nil, nil, nil, nil, nil, nil, nil, nil,
	nil, nil, nil, nil, nil, nil, nil, nil,
	nil, nil, nil, nil, nil, nil, nil, nil,
	nil, nil, nil, nil, nil, nil, nil, halt,
}

func makeOp(code byte, arg uint32, have bool) (op, error) {
	if op := opCodes[code](arg, have); op != nil {
		return op, nil
	}
	return nil, decodeError{code, have, arg}
}

type decodeError struct {
	code byte
	have bool
	arg  uint32
}

func (de decodeError) Error() string {
	return fmt.Sprintf(
		"failed to decode(%02x, %v, %v)",
		de.code, de.arg, de.have)
}
