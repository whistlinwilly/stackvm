package stackvm

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

type op func(*Mach) error

type opDecoder func(arg uint32, have bool) op

var opCodes [128]opDecoder

// TODO: codegen this from the opCodes literal table, rather than building it
// the other way around
const (
	opCodePush   = 0x00
	opCodePop    = 0x01
	opCodeDup    = 0x02
	opCodeSwap   = 0x03
	opCodeFetch  = 0x08
	opCodeStore  = 0x09
	opCodeNeg    = 0x10
	opCodeAdd    = 0x11
	opCodeSub    = 0x12
	opCodeMul    = 0x13
	opCodeDiv    = 0x14
	opCodeMod    = 0x15
	opCodeDivmod = 0x16
	opCodeLt     = 0x18
	opCodeLte    = 0x19
	opCodeEq     = 0x1a
	opCodeNeq    = 0x1b
	opCodeGt     = 0x1c
	opCodeGte    = 0x1d
	opCodeNot    = 0x20
	opCodeAnd    = 0x21
	opCodeOr     = 0x22
	opCodeXor    = 0x23
	opCodeJump   = 0x28
	opCodeJnz    = 0x29
	opCodeJz     = 0x2a
	opCodeMark   = 0x30
	opCodeCall   = 0x31
	opCodeRet    = 0x32
	opCodeCpop   = 0x38
	opCodeP2c    = 0x39
	opCodeC2p    = 0x3a
	opCodeLoop   = 0x40
	opCodeLnz    = 0x41
	opCodeLz     = 0x42
	opCodeFork   = 0x48
	opCodeFnz    = 0x49
	opCodeFz     = 0x4a
	opCodeBranch = 0x50
	opCodeBnz    = 0x51
	opCodeBz     = 0x52
	opCodeHalt   = 0x7f
)

func init() {
	opCodes[opCodePush] = push
	opCodes[opCodePop] = pop
	opCodes[opCodeDup] = dup
	opCodes[opCodeSwap] = swap
	opCodes[opCodeFetch] = fetch
	opCodes[opCodeStore] = store
	opCodes[opCodeNeg] = neg
	opCodes[opCodeAdd] = add
	opCodes[opCodeSub] = sub
	opCodes[opCodeMul] = mul
	opCodes[opCodeDiv] = div
	opCodes[opCodeMod] = mod
	opCodes[opCodeDivmod] = divmod
	opCodes[opCodeLt] = lt
	opCodes[opCodeLte] = lte
	opCodes[opCodeEq] = eq
	opCodes[opCodeNeq] = neq
	opCodes[opCodeGt] = gt
	opCodes[opCodeGte] = gte
	opCodes[opCodeNot] = not
	opCodes[opCodeAnd] = and
	opCodes[opCodeOr] = or
	opCodes[opCodeXor] = xor
	opCodes[opCodeJump] = jump
	opCodes[opCodeJnz] = jnz
	opCodes[opCodeJz] = jz
	opCodes[opCodeMark] = mark
	opCodes[opCodeCall] = call
	opCodes[opCodeRet] = ret
	opCodes[opCodeCpop] = cpop
	opCodes[opCodeP2c] = p2c
	opCodes[opCodeC2p] = c2p
	opCodes[opCodeLoop] = loop
	opCodes[opCodeLnz] = lnz
	opCodes[opCodeLz] = lz
	opCodes[opCodeFork] = fork
	opCodes[opCodeFnz] = fnz
	opCodes[opCodeFz] = fz
	opCodes[opCodeBranch] = branch
	opCodes[opCodeBnz] = bnz
	opCodes[opCodeBz] = bz
	opCodes[opCodeHalt] = halt
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
		"failed to decode(name:%q code:%02x arg:%v have:%v)",
		opCode2Name[de.code],
		de.code, de.arg, de.have)
}
