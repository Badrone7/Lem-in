package main

import (
	"fmt"
	"os"

	z01 "Z01/dependencies"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go <input_file>")
		return
	}
	inputFile := os.Args[1]
	graph, err := z01.ParseInput(inputFile)
	if err != nil {
		fmt.Printf("Error parsing input: %v\n", err)
		return
	}

	paths, err := z01.FindPaths(graph)
	if err != nil {
		fmt.Printf("Error finding paths: %v\n", err)
		return
	}
	if len(paths) == 0 {
		fmt.Println("No paths found from start to end.")
		return
	}

	z01.AssignAntsToPaths(graph, paths)
	allMoves := z01.SimulateAntMovements(graph, paths)
	for _, moves := range allMoves {
		for _, move := range moves {
			fmt.Printf("L%d-%s ", move.AntID, move.RoomName)
		}
		fmt.Println()
	}
}
