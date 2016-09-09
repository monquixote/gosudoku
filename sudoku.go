package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"
)

const dim = 9 //The dimensions of a puzzle
var boxDim = int(math.Sqrt(dim))
var maps = generateBoardMaps()

func main() {
	puzzles := readSudokusFromFile("sudoku.txt")

	//Solving all puzzles serially
	start := time.Now()
	for _, puzzle := range puzzles {
		sets := puzzle2ConstraintSets(puzzle)
		solveBySearch(sets)
	}
	stop := time.Now()
	fmt.Printf("Serial %v \n", stop.Sub(start))

	//Solving all puzzles in parallel
	bools := make(chan bool, len(puzzles))
	start = time.Now()
	for _, puzzle := range puzzles {
		go func(puzzle []int) {
			sets := puzzle2ConstraintSets(puzzle)
			_, res := solveBySearch(sets)
			bools <- res
		}(puzzle)
	}
	for i := 0; i < len(puzzles); i++ {
		<-bools
	}
	stop = time.Now()
	fmt.Printf("Par %v \n", stop.Sub(start))

	sets := puzzle2ConstraintSets(puzzles[0])
	fmt.Println("Before")
	printPuzzle(constraintSet2Puzzle(sets))

	sets, complete := solveBySearch(sets)
	complete, err := checkCompletion(sets)
	if err != nil {
		fmt.Println("Puzzle invalid!")
	}
	if complete {
		fmt.Println("Puzzle solved!")
	} else {
		fmt.Println("Puzzle failed :(")
	}
	printPuzzle(constraintSet2Puzzle(sets))
}

func solveBySearch(sets []map[int]bool) ([]map[int]bool, bool) {
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

func getSearchCandidate(sets []map[int]bool) int {
	for i := 2; i <= dim; i++ {
		for j, elem := range sets {
			if len(elem) == i {
				return j
			}
		}
	}
	return 0
}

func cloneBoard(sets []map[int]bool) []map[int]bool {
	newSet := []map[int]bool{}
	for i, set := range sets {
		newSet = append(newSet, map[int]bool{})
		for key, val := range set {
			newSet[i][key] = val
		}
	}
	return newSet
}

func checkCompletion(sets []map[int]bool) (bool, error) {
	complete := true
	for _, val := range sets {
		if len(val) == 0 {
			return false, errors.New("Invalid sudoku!")
		}
		if len(val) > 1 {
			complete = false
		}
	}
	return complete, nil
}

func printPuzzle(board []int) {
	for i, val := range board {
		fmt.Print(val)
		if (i+1)%dim == 0 {
			fmt.Println()
		}
	}
}

func applyAllConstraints(maps [][][]int, sets []map[int]bool) int {
	changes := 0
	changeChan := make(chan int, dim)
	for _, cat := range maps {
		for _, cMap := range cat {
			go func(sets []map[int]bool, cMap []int) {
				changeChan <- propagateConstraints(sets, cMap)
			}(sets, cMap)
		}
		for i := 0; i < dim; i++ {
			changes += <-changeChan
		}
	}
	return changes
}

func propagateConstraints(sets []map[int]bool, boardMap []int) int {
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

func puzzle2ConstraintSets(puzzle []int) []map[int]bool {
	sets := make([]map[int]bool, len(puzzle))
	for i, val := range puzzle {
		sets[i] = map[int]bool{}
		if val != 0 {
			sets[i][val] = true
		} else {
			for j := 1; j < dim+1; j++ {
				sets[i][j] = true
			}
		}
	}
	return sets
}

func constraintSet2Puzzle(sets []map[int]bool) []int {
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

func readSudokusFromFile(filePath string) [][]int {
	puzzles := make([][]int, 50)
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
	}
	return puzzles
}
