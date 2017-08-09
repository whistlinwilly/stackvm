package stackvm_test

import (
	"testing"

	. "github.com/jcorbin/stackvm/x"
)

// The general idea here is to store the cube as a flat array, generate all
// possible next moves as jumps within this array through some cleverness,
// and otherwise simply loop through a generate(-fork)-check- cycle

// Cleverness:
// Imagine the cube encoded as a flat array in X,Y,Z major order, i.e.
// increasing the index by 1 would be seen simply as moving towards the
// positive X axis in the cube, increasing the index by 3 (in a 3x3 cube)
// is moving in the positive Y axis, and increasing the index by 9 would
// be movement in the positive Z axis. It is easy then to represent a cube
// solution as an array of encoded move choices (1-6) and a "current" position
// (serving double duty during solving) set back to the starting position.

// Generating starting moves:
// Originally I had considered some generative approach to build starting machines
// but most straightforward is a function that just pushes one of the few starting
// positions on the stack and forks to the generate function

// Generating next moves:
// This part is easy, if we have a fixed position in the snake (and not start)
// we propose the last move we made, otherwise we propose one of {1-6}, which
// are indexes in the static "move offset" table (just a level of indirection
// around the set of valid moves {-9, -3, -1, 1, 3, 9})

// Checking valid moves:
// Unfortunately the best scheme I can come up with for checking valid moves is
// to encode all invalid moves for each position as set bits in a 27x6 bit array.
// Then with the stack set up as [proposed move, current position] we can:
// `push 6, mul, add, fetch rel {lookupTable}, hnz`
// (jump to lookup + (6 * current position + proposed move)), halt if not zero
// Then we apply the move to our current index, getting our new index, and now:
// `fetch rel {cubeMemory}, hnz`
// (halt if cell is occupied)
// If these checks pass, we made a valid move into an unoccupied cell, time to
// generate more.

const CUBE_MEM = 0x100

var snakeTest = TestCase{
	Name: "Snake cube solver",
	Prog: MustAssemble(
		// Sample Snake 1 Ascii Art!
		// |X|X|X|
		//     |X|X|
		//       |X|
		//       |X|X|
		//         |X|
		//         |X|X|
		//           |X|X|X|
		//               |X|
		//               |X|X|
		//                 |X|X|
		//                   |X|
		//                   |X|X|X|
		//                       |X|
		//                       |X|X|X|

		// String encoding "fffaaafaafaaafafaaaafafafaf"
		// Bool encoding 000111011011101011110101010

		// Problem / Solution memory is defined over 0x100-0x26c
		//
		// Cube memory 					0x100 - 0x16c [27]uint32	CUBE_MEM
		// Starting position 		0x174 				uint32 			START_POS
		// Snake memory 				0x200 - 0x26c [27]uint32	SNAKE_MEM

		0x40, // stack size

		CUBE_MEM, "cpush", CUBE_MEM+4*(27+1), "cpush", // : 0x0100 0x0174 -- for returning the solution (including starting position)
		":generate_starting_cubes", "call",
		":solve", "call",
		0, "halt",

		// defined functions
		"generate_starting_cubes:", // : retIp
		// starts at (0,0,0)
		0, "push", ":starting_position_set", "fork", // : retIp -- child has 0 on parameter stack
		// starts at (1,0,0)
		1, "push", ":starting_position_set", "fork", // : retIp -- child has 1 on parameter stack
		// starts at (1, 1, 0)
		4, "push", ":starting_position_set", "fork", // : retIp -- child has 4 on parameter stack
		// starts at (1, 1, 1)
		13, "push", ":starting_position_set", "fork", // : retIp -- child has 13 on parameter stack
		"starting_position_set:",
		"ret",

		"solve:",
		"ret",
	),

	Result: Results{
		{Values: generateEmptyValues()},
		{Values: generateEmptyValues()},
		{Values: generateEmptyValues()},
		{Values: generateEmptyValues()},
		{Values: generateEmptyValues()},
	},
}

func generateEmptyValues() [][]uint32 {
	var empty [][]uint32
	var subEmpty []uint32
	for i := 0; i < 28; i++ {
		subEmpty = append(subEmpty, 0)
	}
	empty = append(empty, subEmpty)
	return empty
}

func TestMach_snake_cube(t *testing.T)      { snakeTest.Run(t) }
func BenchmarkMach_snake_cube(b *testing.B) { snakeTest.Bench(b) }
