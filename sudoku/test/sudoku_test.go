package sudoku_test

import (
	"log"
	"os"
	"testing"

	"github.com/monquixote/gosudoku/sudoku"
)

func loadTestFile(fileName string) [][]int {

	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal("Could not open file")
	}
	defer file.Close()

	puzzles, err := sudoku.ReadSudokus(file)

	if err != nil {
		log.Fatal(err)
	}
	return puzzles
}

// Check valid puzzles validate
func TestValidatePuzzleValidCase(t *testing.T) {
	testFiles := []string{"valid_unsolved.txt", "valid_solved.txt"}

	for _, fileName := range testFiles {
		puzzles := loadTestFile(fileName)

		for i, puzzle := range puzzles {
			if !sudoku.ValidatePuzzle(puzzle) {
				t.Errorf("Puzzle %v in %v did not validate", i, fileName)
			}
		}
	}
}

// Check invalid puzzles don't validate
func TestValidatePuzzleInvalidCase(t *testing.T) {
	tests := [][]int{}

	// Size tests
	tests = append(tests, []int{})
	tests = append(tests, []int{0})
	tests = append(tests, make([]int, 9*9+1))
	tests = append(tests, make([]int, 9*9-1))

	// Value tests
	outOfRange1 := make([]int, 9*9)
	outOfRange1[0] = -1
	tests = append(tests, outOfRange1)

	outOfRange2 := make([]int, 9*9)
	outOfRange2[0] = 10
	tests = append(tests, outOfRange2)

	// Duplicate values
	duplicate := make([]int, 9*9)
	duplicate[0] = 1
	duplicate[1] = 1
	tests = append(tests, duplicate)

	for i, puzzle := range tests {
		if sudoku.ValidatePuzzle(puzzle) {
			t.Errorf("Puzzle %v with input %v validated", i, puzzle)
		}
	}
}

func TestSolvePuzzle(t *testing.T) {
	unsolvedPuzzles := loadTestFile("valid_unsolved.txt")
	solvedPuzzles := loadTestFile("valid_solved.txt")

	for i, puzzle := range unsolvedPuzzles {
		candidate, complete := sudoku.SolvePuzzle(puzzle)
		if !complete {
			t.Errorf("Puzzle %v was not solved ", i)
		}
		for j, val := range candidate {
			if val != solvedPuzzles[i][j] {
				t.Errorf("Puzzle %v element %v does not match", i, j)
			}
		}
	}
}

// Sequential Benchmark
func BenchmarkSerial(b *testing.B) {
	puzzles := loadTestFile("../../sudoku.txt")
	b.ResetTimer()

	for _, puzzle := range puzzles {
		sudoku.SolvePuzzle(puzzle)
	}
}

// Parallel Benchmark
func BenchmarkParallel(b *testing.B) {
	puzzles := loadTestFile("../../sudoku.txt")
	b.ResetTimer()

	bools := make(chan bool, len(puzzles))
	for _, puzzle := range puzzles {
		go func(puzzle []int) {
			_, res := sudoku.SolvePuzzle(puzzle)
			bools <- res
		}(puzzle)
	}
	for range puzzles {
		<-bools
	}
}
