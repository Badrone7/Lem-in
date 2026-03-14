package dependencies

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func ParseInput(filename string) (*Graph, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("ERROR: invalid data format, could not open file: %v", err)
	}
	defer file.Close()

	graph := &Graph{
		Rooms: make(map[string]*Room),
	}

	scanner := bufio.NewScanner(file)

	lineNumber := 0
	antsParsed := false
	nextIsStart := false
	nextIsEnd := false
	linksStarted := false

	coordsSeen := make(map[string]bool)
	linksSeen := make(map[string]bool)

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			continue
		}

		// Parse ants first
		if !antsParsed {
			if strings.HasPrefix(line, "#") {
				continue
			}

			ants, err := strconv.Atoi(line)
			if err != nil || ants <= 0 {
				return nil, fmt.Errorf("ERROR: invalid data format, invalid number of ants on line %d", lineNumber)
			}

			graph.Ants = ants
			antsParsed = true
			continue
		}

		// Handle comments and commands
		if strings.HasPrefix(line, "#") {
			switch line {
			case "##start":
				if graph.Start != nil || nextIsStart {
					return nil, fmt.Errorf("ERROR: invalid data format, multiple start commands on line %d", lineNumber)
				}
				if linksStarted {
					return nil, fmt.Errorf("ERROR: invalid data format, start room declared after links on line %d", lineNumber)
				}
				nextIsStart = true
				nextIsEnd = false

			case "##end":
				if graph.End != nil || nextIsEnd {
					return nil, fmt.Errorf("ERROR: invalid data format, multiple end commands on line %d", lineNumber)
				}
				if linksStarted {
					return nil, fmt.Errorf("ERROR: invalid data format, end room declared after links on line %d", lineNumber)
				}
				nextIsEnd = true
				nextIsStart = false

			default:
				// Ignore normal comments and unknown commands
			}
			continue
		}

		// Parse link
		if strings.Contains(line, "-") && !strings.Contains(line, " ") {
			linksStarted = true

			parts := strings.Split(line, "-")
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return nil, fmt.Errorf("ERROR: invalid data format, invalid link definition on line %d", lineNumber)
			}

			name1 := parts[0]
			name2 := parts[1]

			if name1 == name2 {
				return nil, fmt.Errorf("ERROR: invalid data format, self-link on line %d: %s", lineNumber, line)
			}

			room1, ok1 := graph.Rooms[name1]
			room2, ok2 := graph.Rooms[name2]
			if !ok1 || !ok2 {
				return nil, fmt.Errorf("ERROR: invalid data format, link references unknown room on line %d: %s", lineNumber, line)
			}

			linkKey := makeLinkKey(name1, name2)
			if linksSeen[linkKey] {
				return nil, fmt.Errorf("ERROR: invalid data format, duplicate link on line %d: %s", lineNumber, line)
			}

			room1.Links = append(room1.Links, room2)
			room2.Links = append(room2.Links, room1)
			linksSeen[linkKey] = true
			continue
		}

		// Parse room
		fields := strings.Fields(line)
		if len(fields) == 3 {
			if linksStarted {
				return nil, fmt.Errorf("ERROR: invalid data format, room declared after links on line %d", lineNumber)
			}

			name := fields[0]

			if name == "" {
				return nil, fmt.Errorf("ERROR: invalid data format, empty room name on line %d", lineNumber)
			}
			if strings.HasPrefix(name, "L") {
				return nil, fmt.Errorf("ERROR: invalid data format, room name cannot start with L on line %d", lineNumber)
			}
			if strings.HasPrefix(name, "#") {
				return nil, fmt.Errorf("ERROR: invalid data format, room name cannot start with # on line %d", lineNumber)
			}
			if strings.Contains(name, "-") {
				return nil, fmt.Errorf("ERROR: invalid data format, room name cannot contain '-' on line %d", lineNumber)
			}

			x, errX := strconv.Atoi(fields[1])
			y, errY := strconv.Atoi(fields[2])
			if errX != nil || errY != nil {
				return nil, fmt.Errorf("ERROR: invalid data format, invalid coordinates on line %d", lineNumber)
			}

			if _, exists := graph.Rooms[name]; exists {
				return nil, fmt.Errorf("ERROR: invalid data format, duplicate room name on line %d: %s", lineNumber, name)
			}

			coordKey := fmt.Sprintf("%d,%d", x, y)
			if coordsSeen[coordKey] {
				return nil, fmt.Errorf("ERROR: invalid data format, duplicate room coordinates on line %d: %d %d", lineNumber, x, y)
			}

			room := &Room{
				Name:  name,
				X:     x,
				Y:     y,
				Links: []*Room{},
			}

			graph.Rooms[name] = room
			coordsSeen[coordKey] = true

			if nextIsStart {
				graph.Start = room
				nextIsStart = false
			} else if nextIsEnd {
				graph.End = room
				nextIsEnd = false
			}

			continue
		}

		return nil, fmt.Errorf("ERROR: invalid data format, invalid line %d: %s", lineNumber, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("ERROR: invalid data format, reading error: %v", err)
	}

	if !antsParsed || graph.Ants <= 0 {
		return nil, fmt.Errorf("ERROR: invalid data format, invalid number of ants")
	}
	if nextIsStart {
		return nil, fmt.Errorf("ERROR: invalid data format, ##start not followed by a room")
	}
	if nextIsEnd {
		return nil, fmt.Errorf("ERROR: invalid data format, ##end not followed by a room")
	}
	if graph.Start == nil {
		return nil, fmt.Errorf("ERROR: invalid data format, no start room found")
	}
	if graph.End == nil {
		return nil, fmt.Errorf("ERROR: invalid data format, no end room found")
	}

	return graph, nil
}

func makeLinkKey(a, b string) string {
	if a < b {
		return a + "-" + b
	}
	return b + "-" + a
}
