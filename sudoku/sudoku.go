package sudoku

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"math"
	"strconv"
	"strings"
)

const dim = 9 //The dimensions of a puzzle
var boxDim = int(math.Sqrt(dim))

// Values for masking off the board
// 3D slice [type - row/column/box][mask index][mask value]
var masks = generateBoardMasks()

// constraint represents the legal values a square could take
type constraint map[int]bool

func (c *constraint) clone() constraint {
	newc := constraint{}
	for key, val := range *c {
		newc[key] = val
	}
	return newc
}

// constraint factory
// val - puzzle value where 0 represents unknown
func newConstraint(val int) constraint {
	c := constraint{}
	if val != 0 {
		c[val] = true
	} else {
		for i := 1; i < dim+1; i++ {
			c[i] = true
		}
	}
	return c
}

// SolvePuzzle Attempts to solve a sudoku.
// Takes a puzzle as an int slice row by row with 0 representing an unknown value.
// Returns a constraint list of the result of attempting to solve the puzzle and a bool indicating if the attempt to solve succeeded.
func SolvePuzzle(puzzle []int) ([]int, bool) {
	constraints := puzzle2Constraints(puzzle)
	finalSet, solved := solveBySearch(constraints)
	return constraints2Puzzle(finalSet), solved
}

// ReadSudokus Takes sudokus in the Euler 96 text format https://projecteuler.net/problem=96
// Returns a 2D slice containing the parsed puzzles
func ReadSudokus(reader io.Reader) ([][]int, error) {
	puzzles := make([][]int, 50)

	scanner := bufio.NewScanner(reader)
	numberCounter := 0
	puzzleCounter := 0
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "Grid") {
			continue
		}
		for _, runeValue := range scanner.Text() {
			val := int(runeValue - '0')
			if val > 9 || val < 0 {
				return nil, errors.New("Invalid Character in puzzle: " + strconv.Itoa(puzzleCounter) + " element " + strconv.Itoa(numberCounter))
			}
			puzzles[puzzleCounter] = append(puzzles[puzzleCounter], val)
			numberCounter++
		}

		if numberCounter == dim*dim {
			puzzleCounter++
			numberCounter = 0
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if numberCounter != 0 {
		return nil, errors.New("Invalid puzzle dimensions")
	}
	return puzzles, nil
}

// Attempts to solve a sudoku first by propagating constraints and then by brute force search
func solveBySearch(constraints []constraint) ([]constraint, bool) {
	changes := 0
	for changed := true; changed; changed = changes > 0 {
		changes = applyAllConstraints(constraints)
	}
	complete, err := checkCompletion(constraints)
	if err != nil {
		return constraints, false
	}
	if complete {
		return constraints, true
	}

	candidate, err := getSearchCandidate(constraints)
	if err != nil {
		return constraints, false
	}
	for key := range constraints[candidate] {
		clone := cloneBoard(constraints)
		clone[candidate] = newConstraint(key)
		clone, solved := solveBySearch(clone)
		if solved {
			return clone, true
		}
	}

	return constraints, false
}

// Finds the best candidate to use for brute force search favoring more restrictive constraints
func getSearchCandidate(constraints []constraint) (int, error) {
	for i := 2; i <= dim; i++ {
		for j, elem := range constraints {
			if len(elem) == i {
				return j, nil
			}
		}
	}
	return 0, errors.New("No search candidates could be found")
}

// Clones a contraint array
func cloneBoard(constraints []constraint) []constraint {
	cloned := []constraint{}
	for _, c := range constraints {
		cloned = append(cloned, c.clone())
	}
	return cloned
}

// Checks if a puzzle is complete (every square is constrained to a single value)
func checkCompletion(constraints []constraint) (bool, error) {
	for _, val := range constraints {
		if len(val) == 0 {
			return false, errors.New("Invalid sudoku!")
		}
		if len(val) > 1 {
			return false, nil
		}
	}
	return true, nil
}

// Puzzle2String Takes a Puzzle and renders it to a string.
func Puzzle2String(board []int) string {
	var b bytes.Buffer
	for i, val := range board {
		b.WriteString(strconv.Itoa(val))
		if (i+1)%dim == 0 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

// Applies constraints to every row, column and box
// Uses the fact that rows, columns and boxes don't intersect to apply constraints in parallel
func applyAllConstraints(constraints []constraint) int {
	changes := 0
	changeChan := make(chan int, dim)
	for _, category := range masks {
		for _, mask := range category {
			go func(constraints []constraint, mask []int) {
				changeChan <- propagateConstraints(constraints, mask)
			}(constraints, mask)
		}
		for i := 0; i < dim; i++ {
			changes += <-changeChan
		}
	}
	return changes
}

// TODO split into two functions
func propagateConstraints(constraints []constraint, boardMask []int) int {
	return propagateConstraint1(constraints, boardMask) +
		propagateConstraint2(constraints, boardMask)
}

// Propagates constraints using the rule that any known value within a mask cannot appear anywhere else within the mask
// e.g. If you are sure the first element of a row is "1", nothing else in that row could be "1"
func propagateConstraint1(constraints []constraint, boardMask []int) int {
	changes := 0
	for _, cur := range boardMask {
		if len(constraints[cur]) != 1 {
			continue
		}
		for key := range constraints[cur] {
			for _, elem := range boardMask {
				if elem == cur || !constraints[elem][key] {
					continue
				}
				delete(constraints[elem], key)
				changes++
			}
		}
	}

	return changes
}

// Propagate constraints using the rule that if you are the only element
// within your mask permitted to have a value then that must be your value
// e.g. If all of the elements of the first row are forbidden to be 7 except one that element must be 7
func propagateConstraint2(constraints []constraint, boardMask []int) int {
	changes := 0

Outer:
	for i := 1; i <= dim; i++ {
		found := -1
		for _, val := range boardMask {
			if constraints[val][i] {
				if found != -1 || len(constraints[val]) == 1 {
					continue Outer
				} else {
					found = val
				}
			}
		}
		if found > 0 {
			changes += len(constraints[found]) - 1
			constraints[found] = newConstraint(i)
		}
	}
	return changes
}

// Converts a puzzle in euler format to a slice of constraints
func puzzle2Constraints(puzzle []int) []constraint {
	constraints := make([]constraint, len(puzzle))
	for i, val := range puzzle {
		constraints[i] = newConstraint(val)
	}
	return constraints
}

// Converts a slice of constraints into a puzzle in euler format
func constraints2Puzzle(constraints []constraint) []int {
	puzzle := make([]int, len(constraints))
	for i, set := range constraints {
		if len(set) == 1 {
			for key := range set { //TODO A better way?
				puzzle[i] = key
			}
		} else {
			puzzle[i] = 0
		}
	}
	return puzzle
}

// Calculates the values to mask off rows columns and boxes from the puzzle
func generateBoardMasks() [][][]int {
	var rows, cols, boxes [dim][]int

	for i := 0; i < dim; i++ {
		for j := 0; j < dim; j++ {
			rows[i] = append(rows[i], j+i*dim)
			cols[i] = append(cols[i], j*dim+i)
		}
	}
	var offset = 0
	for i := 0; i < dim; i++ {
		for j := 0; j < boxDim; j++ {
			for k := 0; k < boxDim; k++ {
				boxes[i] = append(boxes[i], j+ // shift value horizontal
					(k*dim)+ // shift value vertical
					(i*boxDim)+ // shift box horizontal
					(offset*dim*(boxDim-1))) // Shift box vertical
			}
		}
		if (i+1)%boxDim == 0 {
			offset++
		}
	}
	return [][][]int{rows[:], cols[:], boxes[:]}
}

// ValidatePuzzle determines if a given puzzle is valid
func ValidatePuzzle(puzzle []int) bool {
	if len(puzzle) != dim*dim {
		return false
	}
	for _, maskType := range masks {
		for _, mask := range maskType {
			if !validateMask(puzzle, mask) {
				return false
			}
		}
	}
	return true
}

// Checks for duplicate numbers within a given masked off section of the puzzle
func validateMask(puzzle []int, mask []int) bool {
	seenBefore := make([]bool, dim+1)
	for _, maskVal := range mask {
		if puzzle[maskVal] != 0 && seenBefore[puzzle[maskVal]] == true {
			return false
		}
		seenBefore[puzzle[maskVal]] = true
	}
	return true
}
