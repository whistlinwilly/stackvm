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
// The trickery here is to avoid tedium in generation. It seems to me the
// best solution is to start with the cube array containing 1 in all starting
// positions, then sweep through this array and when we find 1 (fork to sweep)
// set current position to this index before clearing the remaining 1 bits.
// This leaves us with as many machines as we had set bits, where each is
// ready to begin the standard generate(-fork)-check loop.

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

func TestMach_snake_cube(t *testing.T) {
	TestCases{
		{
			Name: "Snake cube solver",
			Err:  "HALT(1)",
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
				0x40,
				1, "halt",
			),
			Result: Result{
				Err: "HALT(1)",
			},
		},
	}.Run(t)
}
