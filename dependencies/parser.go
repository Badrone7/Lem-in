package dependencies

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ParseInput reads the input file and constructs the graph representation of the rooms and links.
func ParseInput(filename string) (*Graph, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("ERROR: could not open file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	graph := &Graph{
		Rooms: make(map[string]*Room),
	}

	lineNumber := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineNumber++

		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if graph.Ants == 0 {
			ants, err := strconv.Atoi(strings.TrimSpace(line))
			if err != nil {
				return nil, fmt.Errorf("ERROR: invalid data format, invalid number of ants on line %d: %v", lineNumber, err)
			}
			graph.Ants = ants
			continue
		}

		if strings.Contains(line, " ") {
			parts := strings.Split(line, " ")
			if len(parts) != 3 {
				return nil, fmt.Errorf("ERROR: invalid data format, invalid room definition on line %d: %s", lineNumber, line)
			}
			name := parts[0]
			x, err1 := strconv.Atoi(parts[1])
			y, err2 := strconv.Atoi(parts[2])
			if err1 != nil || err2 != nil {
				return nil, fmt.Errorf("ERROR: invalid data format, invalid coordinates for room on line %d: %v, %v", lineNumber, err1, err2)
			}
			room := &Room{Name: name, X: x, Y: y}
			graph.Rooms[name] = room
		} else if strings.Contains(line, "-") {
			parts := strings.Split(line, "-")
			if len(parts) != 2 {
				return nil, fmt.Errorf("ERROR: invalid data format, invalid link definition on line %d: %s", lineNumber, line)
			}
			room1, ok1 := graph.Rooms[parts[0]]
			room2, ok2 := graph.Rooms[parts[1]]
			if !ok1 || !ok2 {
				return nil, fmt.Errorf("ERROR: invalid data format, link references undefined room on line %d: %s", lineNumber, line)
			}
			room1.Links = append(room1.Links, room2)
			room2.Links = append(room2.Links, room1)
		} else {
			return nil, fmt.Errorf("ERROR: invalid data format, unrecognized line format on line %d: %s", lineNumber, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("ERROR: error reading file: %v", err)
	}

	return graph, nil
}