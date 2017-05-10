package stackvm

type op func(*Mach) error

type opDecoder func(arg uint32, have bool) op

type opImmKind int

const (
	opImmNone = opImmKind(iota)
	opImmVal
	opImmAddr
	opImmOffset

	opImmType  = 0x0f
	opImmFlags = ^0x0f
	opImmReq   = 0x010
)

func (k opImmKind) kind() opImmKind { return k & opImmType }
func (k opImmKind) required() bool  { return (k & opImmReq) != 0 }

func (k opImmKind) String() string {
	switch k {
	case opImmNone:
		return "NoImmediate"
	case opImmVal:
		return "ImmediateVal"
	case opImmAddr:
		return "ImmediateAddr"
	case opImmOffset:
		return "ImmediateOffset"
	}
	return "InvalidImmediate"
}

type opDef struct {
	name string
	imm  opImmKind
}

var noop = opDef{}

func valop(name string) opDef  { return opDef{name, opImmVal} }
func addrop(name string) opDef { return opDef{name, opImmAddr} }
func offop(name string) opDef  { return opDef{name, opImmOffset} }
func justop(name string) opDef { return opDef{name, opImmNone} }

// TODO: mark required ops
// case opCodePush:
// 	m.err = errImmReq
// case opCodeCpush:
// 	m.err = errImmReq

var ops = [128]opDef{
	// 0x00
	valop("push"), valop("pop"), valop("dup"), valop("swap"),
	noop, noop, noop, noop,
	// 0x08
	addrop("fetch"), addrop("store"),
	noop, noop, noop, noop, noop, noop,
	// 0x10
	valop("add"), valop("sub"),
	valop("mul"), valop("div"),
	valop("mod"), valop("divmod"),
	justop("neg"), noop,
	// 0x18
	valop("lt"), valop("lte"), valop("gt"), valop("gte"),
	valop("eq"), valop("neq"), noop, noop,
	// 0x20
	justop("not"), justop("and"), justop("or"),
	noop, noop, noop, noop, noop,
	// 0x28
	valop("cpush"), valop("cpop"), valop("p2c"), valop("c2p"),
	justop("mark"), noop, noop, noop,
	// 0x30
	offop("jump"), offop("jnz"), offop("jz"),
	justop("loop"), justop("lnz"), justop("lz"),
	addrop("call"), justop("ret"),
	// 0x38
	noop, noop, noop, noop, noop, noop, noop, noop,
	// 0x40
	offop("fork"), offop("fnz"), offop("fz"),
	noop, noop, noop, noop, noop,
	// 0x48
	noop, noop, noop, noop, noop, noop, noop, noop,
	// 0x50
	offop("branch"), offop("bnz"), offop("bz"),
	noop, noop, noop, noop, noop,
	// 0x58
	noop, noop, noop, noop, noop, noop, noop, noop,
	// 0x60
	noop, noop, noop, noop, noop, noop, noop, noop,
	// 0x68
	noop, noop, noop, noop, noop, noop, noop, noop,
	// 0x70
	noop, noop, noop, noop, noop, noop, noop, noop,
	// 0x78
	noop, noop, noop, noop, noop,
	valop("hnz"), valop("hz"), valop("halt"),
}

const (
	// TODO: codegen this
	opCodePush   = 0x00
	opCodePop    = 0x01
	opCodeDup    = 0x02
	opCodeSwap   = 0x03
	opCodeFetch  = 0x08
	opCodeStore  = 0x09
	opCodeAdd    = 0x10
	opCodeSub    = 0x11
	opCodeMul    = 0x12
	opCodeDiv    = 0x13
	opCodeMod    = 0x14
	opCodeDivmod = 0x15
	opCodeNeg    = 0x16
	opCodeLt     = 0x18
	opCodeLte    = 0x19
	opCodeGt     = 0x1a
	opCodeGte    = 0x1b
	opCodeEq     = 0x1c
	opCodeNeq    = 0x1d
	opCodeNot    = 0x20
	opCodeAnd    = 0x21
	opCodeOr     = 0x22
	opCodeCpush  = 0x28
	opCodeCpop   = 0x29
	opCodeP2c    = 0x2a
	opCodeC2p    = 0x2b
	opCodeMark   = 0x2c
	opCodeJump   = 0x30
	opCodeJnz    = 0x31
	opCodeJz     = 0x32
	opCodeLoop   = 0x33
	opCodeLnz    = 0x34
	opCodeLz     = 0x35
	opCodeCall   = 0x36
	opCodeRet    = 0x37
	opCodeFork   = 0x40
	opCodeFnz    = 0x41
	opCodeFz     = 0x42
	opCodeBranch = 0x50
	opCodeBnz    = 0x51
	opCodeBz     = 0x52
	opCodeHnz    = 0x7d
	opCodeHz     = 0x7e
	opCodeHalt   = 0x7f
)

var opName2Code = make(map[string]byte, 128)

func init() {
	for i, def := range ops {
		opName2Code[def.name] = byte(i)
	}
}
