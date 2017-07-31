package stackvm_test

import (
	"fmt"
	"strings"
	"testing"
)

// . "github.com/jcorbin/stackvm/x"

func genSnakeCubeRows(rng fastRNG, m int) []int {
	n := m * m * m
	r := make([]int, 0, n)
	i := 0
	run := 0
	for i < n {
		var c int
		for {
			c = 1 + int(rng.next()%3)
			if i+c > n {
				continue
			}
			if c == 1 {
				if run >= 3 {
					continue
				}
				run++
			} else {
				run = 2
			}
			break
		}
		i += c
		r = append(r, c)
	}
	return r
}

type cellLabel uint8
type rowLabel []cellLabel

const (
	fixedCell cellLabel = 0
	rowHead   cellLabel = 1 << iota
	rowTail
	colHead
	colTail
)

func (cl cellLabel) String() string {
	if cl == fixedCell {
		return "#"
	}

	parts := make([]string, 0, 6)

	switch cl & (rowHead | rowTail) {
	case rowHead:
		parts = append(parts, "rH")
		cl &= ^rowHead
	case rowTail:
		parts = append(parts, "rT")
		cl &= ^rowTail
	}

	switch cl & (colHead | colTail) {
	case colHead:
		parts = append(parts, "cH")
		cl &= ^colHead
	case colTail:
		parts = append(parts, "cT")
		cl &= ^colTail
	}

	if cl != 0 {
		return fmt.Sprintf("!<%d>!", cl)
	}

	return strings.Join(parts, ":")
}

// labelrows generates a list of row labels given a list of row counts.
//
// rows is simply a list of cell counts per row that describes a possible snake
// (its ability to actually form a cube is another matter). For example,
// consider the trivial 2x2x2 cube, one of the few possible snakes would be [2,
// 1, 2, 1, 2], which can be visualized like:
//  # #
// 	  #
// 	  # #
// 		#
// 		# #
//
// The labels emitted are one of:
// - rH / rT : the cell is the head or tail of a row freedom
// - cH / cT : the cell is the head or tail of a column freedom
// - #       : the cell is not part of a freedom
func labelrows(rows []int) []rowLabel {
	n := 0
	for _, row := range rows {
		n += row
	}
	r := make([]rowLabel, 0, n)

	var tail *cellLabel

	for i, row := range rows {
		rl := make(rowLabel, row)

		if tail != nil && (row > 1 || i == len(rows)-1) {
			addLabel(tail, colHead)
			addLabel(&rl[0], colTail)
		}

		if row > 1 {
			addLabel(&rl[0], rowHead)
			addLabel(&rl[row-1], rowTail)

			tail = &rl[row-1]
		}

		r = append(r, rl)
	}

	return r
}

func addLabel(cl *cellLabel, l cellLabel) {
	if *cl == fixedCell {
		*cl = l
		return
	}
	*cl |= l
}

func renderRowLabels(rls []rowLabel) [][]string {
	r := make([][]string, len(rls))
	for i, rl := range rls {
		ri := make([]string, len(rl))
		for j, cl := range rl {
			ri[j] = cl.String()
		}
		r[i] = ri
	}
	return r
}

// padRowLabels pads initial and final labels within each row label so that
// they will right-align when stacked vertically (next head under prior tail).
func padRowLabels(rowlabels [][]string) {
	var (
		w    int
		last []string
	)
	for _, rl := range rowlabels {
		if len(rl[0]) < w {
			rl[0] = strings.Repeat(" ", w-len(rl[0])) + rl[0]
		}
		if w > 0 && w < len(rl[0]) {
			last[len(last)-1] = strings.Repeat(" ", len(rl[0])-w) + last[len(last)-1]
		}
		w = len(rl[len(rl)-1])
		last = rl
	}
}

func Test_genSnakeCubeRows(t *testing.T) {
	// XXX temp workspace
	rng := makeFastRNG(15517)

	for i := 0; i < 4; i++ {
		rows := genSnakeCubeRows(rng, 3)
		fmt.Println(rows)

		rowlabels := labelrows(rows)
		strRowLabels := renderRowLabels(rowlabels)
		padRowLabels(strRowLabels)

		var prefix string
		for i, row := range rows {
			rl := strRowLabels[i]
			label := strings.Join(rl, " ")
			fmt.Printf("%v: %s%s\n", row, prefix, label)
			prefix += strings.Repeat(" ", len(label)-len(rl[len(rl)-1]))
		}

		fmt.Println()
	}
}

/*

var snakeSolTest = TestCase{
	Name: "snake XXX",
	Prog: MustAssemble(
		0x40,
		//// Done
		0, "halt",
	),
	// Result: XXX,
}

func genSnakeCube(rng uint32, m int) []bool {
	n := m * m * m
	limit := n * n

reboot:
	// fmt.Printf("reboot!\n")
	freedoms := make([]bool, 0, n)
	oc := make([]bool, n)
	var lastPos, lastDir [3]int

	oc[0] = true
	freedoms = append(freedoms, true)

	i := 0
	c := 0
	for {
		if c++; c > limit {
			goto reboot
		}
		rng = lcgM*rng + lcgC

		pos := lastPos
		delta := 1
		if rng%2 == 1 {
			delta = -1
		}
		pos[rng%3] += delta

		if pos[0] < 0 || pos[0] >= m || pos[1] < 0 || pos[1] >= m || pos[2] < 0 || pos[2] >= m {
			continue
		}

		j := m*m*pos[0] + m*pos[1] + pos[2]
		if !oc[j] {
			// fmt.Printf("bop [%v] %v\n", i, pos)
			// TODO: trace these choices to build the snake

			dir := [3]int{pos[0] - lastPos[0], pos[1] - lastPos[1], pos[2] - lastPos[2]}
			if lastDir == [3]int{0, 0, 0} {
				freedoms = append(freedoms, false)
			} else if dir == lastDir {
				freedoms = append(freedoms, false)
			} else {
				freedoms = append(freedoms, true)
			}

			lastDir, lastPos, oc[j] = dir, pos, true
			if i++; i >= n-1 {
				break
			}
		}
	}

	// fmt.Printf("done!\n")

	if cap(freedoms) != len(freedoms) {
		panic("genSnakeCube failed to generate a complete snake cube")
	}

	return freedoms
}


*/
type fastRNG struct{ state *uint32 }

func makeFastRNG(seed uint32) fastRNG { return fastRNG{state: &seed} }

func (fr fastRNG) next() uint32 {
	const (
		M = 134775813
		C = 1
	)
	n := *fr.state
	n = M*n + C
	*fr.state = n
	return n
}
