package xstackvm

import (
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/jcorbin/stackvm"
)

// Assemble builds a byte encoded machine program from a slice of
// operation names. Operations may be preceded by an immediate
// argument. An immediate argument may be an integer value, or a label
// reference string of the form ":name". Labels are defined with a string of
// the form "name:".
func Assemble(in ...interface{}) ([]byte, error) {
	if len(in) < 2 {
		return nil, errors.New("program too short, need at least options and one token")
	}

	// first element is ~ machine options
	var opts stackvm.MachOptions
	switch v := in[0].(type) {
	case int:
		if v < +0 || v > 0xffff {
			return nil, fmt.Errorf("stackSize %d out of range, must be in (0, 65536)", v)
		}
		opts.StackSize = uint16(v)

	case stackvm.MachOptions:
		opts = v

	default:
		return nil, fmt.Errorf("invalid machine options, "+
			"expected a stackvm.MachOptions or an int, "+
			"but got %T(%v) instead",
			v, v)
	}

	// rest is tokens
	toks, err := tokenize(in[1:])
	if err != nil {
		return nil, err
	}

	ops, jumps, err := resolve(toks)
	if err != nil {
		return nil, err
	}

	return assemble(opts, ops, jumps), nil
}

// Alloc can be used as an assembly directive in the ".data" section, where it
// will be expanded to n-many immediate 0s.
type Alloc uint

// MustAssemble uses assemble the input, using Assemble(), and panics
// if it returns a non-nil error.
func MustAssemble(in ...interface{}) []byte {
	prog, err := Assemble(in...)
	if err != nil {
		panic(err)
	}
	return prog
}

type tokenType uint8

const (
	dataSectionTokenType tokenType = iota + 1
	textSectionTokenType
	labelToken
	refToken
	opToken
	immToken
	dataToken
)

func (tt tokenType) String() string {
	switch tt {
	case dataSectionTokenType:
		return ".data"
	case textSectionTokenType:
		return ".text"
	case labelToken:
		return "label"
	case refToken:
		return "ref"
	case opToken:
		return "op"
	case immToken:
		return "imm"
	case dataToken:
		return "data"
	default:
		return fmt.Sprintf("InvalidTokenType(%d)", tt)
	}
}

type token struct {
	t tokenType
	s string
	d uint32
}

var (
	dataSectionToken = token{t: dataSectionTokenType}
	textSectionToken = token{t: textSectionTokenType}
)

func label(s string) token  { return token{t: labelToken, s: s} }
func ref(s string) token    { return token{t: refToken, s: s} }
func opName(s string) token { return token{t: opToken, s: s} }
func imm(n int) token       { return token{t: immToken, d: uint32(n)} }
func data(d uint32)         { return token{t: dataToken, d: d} }

func (t token) String() string {
	switch t.t {
	case labelToken:
		return t.s + ":"
	case refToken:
		return ":" + t.s
	case opToken:
		return t.s
	case immToken:
		return strconv.Itoa(int(t.d))
	default:
		return fmt.Sprintf("InvalidToken(t:%d, s:%q, d:%v)", t.t, t.s, t.d)
	}
}

func tokenize(in []interface{}) (out []token, err error) {
	i := 0

	goto text

data:
	out = append(out, dataSectionToken)
	for ; i < len(in); i++ {
		if s, ok := in[i].(string); ok {
			// directive
			if len(s) > 1 && s[0] == '.' {
				switch s[1:] {
				case "data":
					continue
				case "text":
					goto text
				default:
					return nil, fmt.Errorf("invalid directive %s", s)
				}
			}

			// label
			if j := len(s) - 1; j > 0 && s[j] == ':' {
				out = append(out, label(s[:j]))
				continue
			}

			return nil, fmt.Errorf("unexpected string %q", s)
		}

		// alloc
		if n, ok := in[i].(Alloc); ok {
			for i := 0; i < n; i++ {
				out = append(out, data(0))
			}
			continue
		}

		// data word
		if n, ok := in[i].(int); ok {
			out = append(out, data(n))
			continue
		}

		return nil, fmt.Errorf(
			`invalid token %T(%v); expected ".directive", "label:", or an int`,
			in[i], in[i])
	}

text:
	out = append(out, textSectionToken)
	for ; i < len(in); i++ {
		if s, ok := in[i].(string); ok {
			// directive
			if len(s) > 1 && s[0] == '.' {
				switch s[1:] {
				case "data":
					goto data
				case "text":
					continue
				default:
					return nil, fmt.Errorf("invalid directive %s", s)
				}
			}

			// label
			if j := len(s) - 1; j > 0 && s[j] == ':' {
				out = append(out, label(s[:j]))
				continue
			}

			// ref
			if len(s) > 1 && s[0] == ':' {
				out = append(out, ref(s[1:]))
				goto op
			}

			// opName
			out = append(out, opName(s))
			continue
		}

		// imm
		if n, ok := in[i].(int); ok {
			out = append(out, imm(n))
			goto op
		}

		return nil, fmt.Errorf(
			`invalid token %T(%v); expected ".directive", "label:", ":ref", "opName", or an int`,
			in[i], in[i])

	op:
		i++
		// got r ref or v imm, must have opName
		if s, ok := in[i].(string); ok {
			out = append(out, opName(s))
			continue
		}

		return nil, fmt.Errorf(
			`invalid token %T(%v); expected "opName"`,
			in[i], in[i])
	}
	return
}

func resolve(toks []token) (ops []stackvm.Op, jumps []int, err error) {
	numJumps := 0
	labels := make(map[string]int)
	refs := make(map[string][]int)

	for i := 0; i < len(toks); i++ {
		switch tok := toks[i]; tok.t {
		case textSectionTokenType:
			continue
		case dataSectionTokenType:
			return nil, nil, fmt.Errorf("data section unimplemented")

		case labelToken:
			labels[tok.s] = len(ops)

		case refToken:
			ref := tok.s
			// resolve label references
			i++
			tok = toks[i]
			if tok.t != opToken {
				return nil, nil, fmt.Errorf("next token must be an op, got %v instead", tok.t)
			}
			op, err := stackvm.ResolveOp(tok.s, 0, true)
			if err != nil {
				return nil, nil, err
			}
			if !op.AcceptsRef() {
				return nil, nil, fmt.Errorf("%v does not accept ref %q", op, ref)
			}
			ops = append(ops, op)
			refs[ref] = append(refs[ref], len(ops)-1)
			numJumps++

		case opToken:
			// op without immediate arg
			op, err := stackvm.ResolveOp(tok.s, 0, false)
			if err != nil {
				return nil, nil, err
			}
			ops = append(ops, op)

		case immToken:
			// op with immediate arg
			arg := tok.d
			i++
			tok = toks[i]
			if tok.t != opToken {
				return nil, nil, fmt.Errorf("next token must be an op, got %v instead", tok.t)
			}

			op, err := stackvm.ResolveOp(tok.s, arg, true)
			if err != nil {
				return nil, nil, err
			}
			ops = append(ops, op)

		default:
			return nil, nil, fmt.Errorf("unexpected %v token", tok.t)
		}
	}

	if numJumps > 0 {
		jumps = make([]int, 0, numJumps)
		for name, sites := range refs {
			i, ok := labels[name]
			if !ok {
				return nil, nil, fmt.Errorf("undefined label %q", name)
			}
			for _, j := range sites {
				ops[j].Arg = uint32(i - j - 1)
				jumps = append(jumps, j)
			}
		}
	}

	return
}

type jumpCursor struct {
	jumps []int // op indices that are jumps
	offs  []int // jump offsets, mined out of op args
	i     int   // index of the current jump in jumps...
	ji    int   // ...op index of its jump
	ti    int   // ...op index of its target
}

func makeJumpCursor(ops []stackvm.Op, jumps []int) jumpCursor {
	sort.Ints(jumps)
	jc := jumpCursor{jumps: jumps, ji: -1, ti: -1}
	if len(jumps) > 0 {
		// TODO: offs only for jumps
		offs := make([]int, len(ops))
		for i := range ops {
			offs[i] = int(int32(ops[i].Arg))
		}
		jc.offs = offs
		jc.ji = jc.jumps[0]
		jc.ti = jc.ji + 1 + jc.offs[jc.ji]
	}
	return jc
}

func (jc jumpCursor) next() jumpCursor {
	jc.i++
	if jc.i >= len(jc.jumps) {
		jc.ji, jc.ti = -1, -1
	} else {
		jc.ji = jc.jumps[jc.i]
		jc.ti = jc.ji + 1 + jc.offs[jc.ji]
	}
	return jc
}

func (jc jumpCursor) rewind(ri int) jumpCursor {
	for i, ji := range jc.jumps {
		ti := ji + 1 + jc.offs[ji]
		if ji >= ri || ti >= ri {
			jc.i, jc.ji, jc.ti = i, ji, ti
			break
		}
	}
	return jc
}

func assemble(opts stackvm.MachOptions, ops []stackvm.Op, jumps []int) []byte {
	// setup jump tracking state
	jc := makeJumpCursor(ops, jumps)

	// allocate worst-case-estimated output space
	est, ejc := 0, jc
	for i := range ops {
		if i == ejc.ji {
			est += 5
			ejc = ejc.next()
		} else if ops[i].Have {
			est += ops[i].NeededSize()
		}
		est++
	}

	buf := make([]byte, est+5)
	n := opts.EncodeInto(buf)
	assembleInto(opts, ops, jc, buf[n:])
	return buf
}

func assembleInto(opts stackvm.MachOptions, ops []stackvm.Op, jc jumpCursor, p []byte) []byte {
	base := uint32(opts.StackSize)
	offsets := make([]uint32, len(ops)+1)
	c, i := uint32(0), 0 // current op offset and index
	for i < len(ops) {
		// fix a previously encoded jump's target
		for 0 <= jc.ji && jc.ji < i && jc.ti <= i {
			jIP := base + offsets[jc.ji]
			tIP := base
			if jc.ti < i {
				tIP += offsets[jc.ti]
			} else { // jc.ti == i
				tIP += c
			}
			ops[jc.ji] = ops[jc.ji].ResolveRefArg(jIP, tIP)
			// re-encode the jump and rewind if arg size changed
			lo, hi := offsets[jc.ji], offsets[jc.ji+1]
			if end := lo + uint32(ops[jc.ji].EncodeInto(p[lo:])); end != hi {
				i, c = jc.ji+1, end
				offsets[i] = c
				jc = jc.rewind(i)
			} else {
				jc = jc.next()
			}
		}
		// encode next operation
		c += uint32(ops[i].EncodeInto(p[c:]))
		i++
		offsets[i] = c
	}
	return p[:c]
}
