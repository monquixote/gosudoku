package sudoku

import (
	"bufio"
	"bytes"
	"errors"
	"math"
	"os"
	"strconv"
	"strings"
)

//constraintList Represents the constraints on the sudoku state
//For each square map keys indicate which values are still legal moves
type constraintList []constraint

type constraint map[int]bool

func (c *constraint) clone() constraint {
	newc := map[int]bool{}
	for key, val := range *c {
		newc[key] = val
	}
	return newc
}

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

const dim = 9 //The dimensions of a puzzle
var boxDim = int(math.Sqrt(dim))
var maps = generateBoardMaps()

//SolvePuzzle Attempts to solve a sudoku.
//Takes a puzzle as an int slice row by row with 0 representing unknown values.
//Returns a constraint list of the result of attempting to solve the puzzle and a bool indicating if the attempt to solve succeeded.
func SolvePuzzle(puzzle []int) ([]int, bool) {
	sets := puzzle2ConstraintSets(puzzle)
	finalSet, solved := solveBySearch(sets)
	return constraintSet2Puzzle(finalSet), solved
}

//ReadSudokusFromFile Takes sudokus in the Euler 96 text format https://projecteuler.net/problem=96
//Returns a 2D slice containing the parsed puzzles
func ReadSudokusFromFile(filePath string) ([][]int, error) {
	puzzles := make([][]int, 50) //TODO don't make this fixed to 50 puzzles
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	numberCounter := 0
	puzzleCounter := 0
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "Grid") {
			continue
		}
		for _, runeValue := range scanner.Text() {
			puzzles[puzzleCounter] = append(puzzles[puzzleCounter], int(runeValue-'0'))
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
	return puzzles, nil
}

func solveBySearch(sets constraintList) (constraintList, bool) {
	changes := 0
	for ok := true; ok; ok = changes > 0 {
		changes = applyAllConstraints(maps, sets)
	}
	complete, err := checkCompletion(sets)
	if err != nil {
		return sets, false
	}
	if complete {
		return sets, true
	}

	candidate := getSearchCandidate(sets)
	for key := range sets[candidate] {
		clone := cloneBoard(sets)
		clone[candidate] = map[int]bool{key: true}
		clone, solved := solveBySearch(clone)
		if solved {
			return clone, true
		}
	}

	return sets, false
}

func getSearchCandidate(sets constraintList) int {
	for i := 2; i <= dim; i++ {
		for j, elem := range sets {
			if len(elem) == i {
				return j
			}
		}
	}
	return 0
}

func cloneBoard(sets constraintList) constraintList {
	newSet := constraintList{}
	for _, set := range sets {
		newSet = append(newSet, set.clone())
	}
	return newSet
}

func checkCompletion(sets constraintList) (bool, error) {
	for _, val := range sets {
		if len(val) == 0 {
			return false, errors.New("Invalid sudoku!")
		}
		if len(val) > 1 {
			return false, nil
		}
	}
	return true, nil
}

//Puzzle2String Takes a Puzzle and prints it.
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

func applyAllConstraints(maps [][][]int, sets constraintList) int {
	changes := 0
	changeChan := make(chan int, dim)
	for _, cat := range maps {
		for _, cMap := range cat {
			go func(sets constraintList, cMap []int) {
				changeChan <- propagateConstraints(sets, cMap)
			}(sets, cMap)
		}
		for i := 0; i < dim; i++ {
			changes += <-changeChan
		}
	}
	return changes
}

func propagateConstraints(sets constraintList, boardMap []int) int {
	changes := 0
	for _, cur := range boardMap {
		if len(sets[cur]) != 1 {
			continue
		}
		for key := range sets[cur] {
			for _, elem := range boardMap {
				if elem == cur || !sets[elem][key] {
					continue
				}
				delete(sets[elem], key)
				changes++
			}
		}
	}

	//Constraint type 2
Outer:
	for i := 1; i <= dim; i++ {
		found := -1
		for _, val := range boardMap {
			if sets[val][i] {
				if found != -1 || len(sets[val]) == 1 {
					continue Outer
				} else {
					found = val
				}
			}
		}
		if found > 0 {
			changes += len(sets[found]) - 1
			sets[found] = map[int]bool{i: true}
		}
	}
	return changes
}

func puzzle2ConstraintSets(puzzle []int) constraintList {
	sets := make(constraintList, len(puzzle))
	for i, val := range puzzle {
		sets[i] = newConstraint(val)
	}
	return sets
}

func constraintSet2Puzzle(sets constraintList) []int {
	puzzle := make([]int, len(sets))
	for i, set := range sets {
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

func generateBoardMaps() [][][]int {
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
				boxes[i] = append(boxes[i], j+ //shift value horizontal
					(k*dim)+ //shift value vertical
					(i*boxDim)+ //shift box horizontal
					(offset*dim*(boxDim-1))) //Shift box vertical
			}
		}
		if (i+1)%boxDim == 0 {
			offset++
		}
	}
	return [][][]int{rows[:], cols[:], boxes[:]}
}
