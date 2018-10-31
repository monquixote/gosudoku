package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/monquixote/gosudoku/sudoku"
)

func main() {

	args := os.Args[1:]
	if len(args) == 0 {
		log.Fatal("Filename required as first argument")
	}

	file, err := os.Open(args[0])

	if err != nil {
		log.Fatal("Could not open file")
	}
	defer file.Close()

	puzzles, err := sudoku.ReadSudokus(file)

	if err != nil {
		log.Fatal(err)
	}

	for i, puzzle := range puzzles {
		if !sudoku.ValidatePuzzle(puzzle) {
			log.Fatalf("Puzzle import failed puzzle %d invalid ", i)
		}
	}

	//Solving all puzzles serially
	start := time.Now()
	for _, puzzle := range puzzles {
		solvedPuzzle, solved := sudoku.SolvePuzzle(puzzle)
		if solved == false {
			log.Fatal("Failed to solve puzzle")
		}

		fmt.Print(sudoku.Puzzle2String(solvedPuzzle) + "\n")

		if !sudoku.ValidatePuzzle(solvedPuzzle) {
			log.Fatal("Solution Invalid")
		}
	}
	stop := time.Now()
	fmt.Printf("Serial %v \n", stop.Sub(start))

	//Solving all puzzles in parallel
	bools := make(chan bool, len(puzzles))
	start = time.Now()
	for _, puzzle := range puzzles {
		go func(puzzle []int) {
			_, res := sudoku.SolvePuzzle(puzzle)
			bools <- res
		}(puzzle)
	}
	for i := 0; i < len(puzzles); i++ {
		<-bools
	}
	stop = time.Now()
	fmt.Printf("Par %v \n", stop.Sub(start))

	fmt.Println("Before")
	fmt.Println(sudoku.Puzzle2String(puzzles[0]))

	constraints, complete := sudoku.SolvePuzzle(puzzles[0])

	if complete {
		fmt.Println("Puzzle solved!")
	} else {
		fmt.Println("Puzzle failed :(")
	}
	fmt.Println(sudoku.Puzzle2String(constraints))
}
