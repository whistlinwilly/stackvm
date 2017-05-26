package stackvm_test

import (
	"fmt"
	"testing"

	. "github.com/jcorbin/stackvm/x"
)

// TODO use:
// - mark / loop et al
// - mark / ret (fbret et al)

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
				"storeTo", // v : i

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
				"p2c",     // v v i : i=i+4
				"storeTo", // v : i
				"dup",     // v v : i
				1, "eq",   // v v==1 : i
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

var collatzExplore = TestCase{
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
		"swap",    // v v i d : i
		"p2c",     // v v i : i d
		"storeTo", // v : i d

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
}

func TestMach_collatz_explore(t *testing.T)      { collatzExplore.Run(t) }
func BenchmarkMach_collatz_explore(b *testing.B) { collatzExplore.Bench(b) }

var smmTest = TestCase{
	Name: "send more money (bottom up)",
	Prog: MustAssemble(
		0x40,

		//     s e n d
		// +   m o r e
		// -----------
		//   m o n e y

		// used   [10]uint32 @0x0100    TODO use a bit vector
		// values [8]uint32  @0x0140
		//                   0 1 2 3 4 5 6 7
		//                   d e y n r o s m

		//// d + e = y  (mod 10)

		0x0140, "cpush", 0x0140+4*8, "cpush", // : 0x0140 0x0160

		0x0140+4*0, "push", ":choose", "call", // $d :
		0x0140+4*1, "push", ":choose", "call", // $d $e :
		"add", "dup", // $d+e $d+e :
		10, "mod", // $d+e ($d+e)%10 :
		"dup", 0x0140+4*2, "storeTo", // $d+e $y :   -- $y=($d+e)%10
		":markUsed", "call", // $d+e :
		10, "div", // carry :

		//// carry + n + r = e  (mod 10)

		"dup",               // carry carry :
		0x0140+4*1, "fetch", // carry carry $e :
		"swap",                                // carry $e carry :
		0x0140+4*3, "push", ":choose", "call", // carry $e carry $n :
		"add", "sub", 10, "mod", // carry ($e-(carry+$n))%10 :
		"dup", 0x0140+4*4, "storeTo", // carry $r :   -- $r=($e-(carry+$n))%10
		":markUsed", "call", // carry :
		0x0140+4*3, "fetch", // carry $n :
		0x0140+4*4, "fetch", // carry $n $r :
		"add", "add", 10, "div", // carry :

		//// carry + e + o = n  (mod 10)

		"dup",               // carry carry :
		0x0140+4*1, "fetch", // carry carry $e :
		"add",               // carry carry+$e :
		0x0140+4*3, "fetch", // carry carry+$e $n :
		"swap", "sub", // carry $n-(carry+$e) :
		10, "mod", // carry ($n-(carry+$e))%10 :
		"dup", 0x0140+4*5, "storeTo", // carry $o :   -- $o=($n-(carry+$e))%10
		":markUsed", "call", // carry :
		0x0140+4*1, "fetch", // carry $e :
		0x0140+4*5, "fetch", // carry $e $o :
		"add", "add", 10, "div", // carry :

		//// carry + s + m = o  (mod 10)

		"dup",                                 // carry carry :
		0x0140+4*6, "push", ":choose", "call", // carry carry $s :
		"add",               // carry carry+$s :
		0x0140+4*5, "fetch", // carry carry+$s $o :
		"swap", "sub", // carry $o-(carry+$s) :
		10, "mod", // carry ($o-(carry+$s))%10 :
		"dup", 0x0140+4*7, "storeTo", // carry $m :   -- $m=($o-(carry+$s))%10
		":markUsed", "call", // carry :
		0x0140+4*6, "fetch", // carry $s :
		"dup", 1, "hz", // carry $s :   -- guard $s != 0
		0x0140+4*7, "fetch", // carry $s $m :
		"dup", 1, "hz", // carry $s $m :   -- guard $m != 0
		"add", "add", 10, "div", // carry :

		//// carry = m  (mod 10)
		0x0140+4*7, "fetch", // carry $m
		"eq", 3, "hz",

		//// Done
		0, "halt",

		"choose:",                        // &$X : retIp
		0, "push", ":chooseLoop", "jump", // &$X i=0 : retIp
		"chooseNext:", 1, "add", // &$X i++ : retIp
		"chooseLoop:",  // &$X i : retIp
		"dup", 9, "lt", // &$X i i<9 : retIp
		":chooseNext", "fnz", // &$X i : retIp
		"dup", 4, "mul", 0x0100, "add", // &$X i &used[i] : retIp
		"dup", "fetch", // &$X i &used[i] used[i] : retIp
		2, "hnz", // &$X i &used[i] : retIp
		1, "store", // &$X i : retIp -- used[i]=1
		"dup", 2, "swap", "storeTo", // $X=i : retIp
		"ret", // $X :

		"markUsed:",             // $X : retIp
		4, "mul", 0x0100, "add", // ... &used[$X]
		"dup", "fetch", // ... &used[$X] used[$X]
		2, "hnz", // ... &used[$X]
		1, "store", // ... -- used[$X] = 1
		"ret", // :

	),

	Result: Results{
		{Values: [][]uint32{{
			7, // d
			5, // e
			2, // y
			6, // n
			8, // r
			0, // o
			9, // s
			1, // m
		}}},
	}.WithExpectedHaltCodes(1, 2, 3),
}

func TestMach_send_more_money(t *testing.T)      { smmTest.Run(t) }
func BenchmarkMach_send_more_money(b *testing.B) { smmTest.Bench(b) }
