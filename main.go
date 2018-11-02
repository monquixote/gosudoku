package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/monquixote/gosudoku/sudoku"
)

func main() {
	sudokuPath := flag.String("file", "sudoku.txt", "Path to the file containing sudokus")
	flag.Parse()

	file, err := os.Open(*sudokuPath)

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

	//Solving all puzzles
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
	fmt.Printf("All puzzles solved in %v \n", stop.Sub(start))
}
