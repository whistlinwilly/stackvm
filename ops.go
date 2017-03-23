package stackvm

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

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

var (
	opName2Code = make(map[string]byte, 128)
	opCode2Name [128]string
)

func init() {
	for i, op := range opCodes {
		code := byte(i)
		pc := reflect.ValueOf(op).Pointer()
		f := runtime.FuncForPC(pc)
		name := f.Name()
		if j := strings.LastIndex(name, "."); j >= 0 {
			name = name[j+1:]
		}
		opName2Code[name] = code
		opCode2Name[code] = name
	}
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
