package dependencies

import "errors"

func FindPaths(graph *Graph) ([]*Path, error) {
	if graph == nil {
		return nil, errors.New("graph is nil")
	}
	if graph.Start == nil || graph.End == nil {
		return nil, errors.New("start or end room not defined")
	}

	queue := [][]*Room{{graph.Start}}

	var shortestPaths [][]*Room
	shortestLength := -1

	for len(queue) > 0 {
		currentPath := queue[0]
		queue = queue[1:]

		lastRoom := currentPath[len(currentPath)-1]

		// Stop exploring longer paths once shortest found
		if shortestLength != -1 && len(currentPath) > shortestLength {
			continue
		}

		if lastRoom == graph.End {
			shortestLength = len(currentPath)
			shortestPaths = append(shortestPaths, currentPath)
			continue
		}

		for _, neighbor := range lastRoom.Links {
			if !containsRoom(currentPath, neighbor) {
				newPath := append([]*Room{}, currentPath...)
				newPath = append(newPath, neighbor)
				queue = append(queue, newPath)
			}
		}
	}

	if len(shortestPaths) == 0 {
		return nil, errors.New("no path found")
	}

	var paths []*Path
	for _, rooms := range shortestPaths {
		paths = append(paths, &Path{
			Rooms:        rooms,
			AntsAssigned: 0,
		})
	}

	return paths, nil
}

func containsRoom(path []*Room, room *Room) bool {
	for _, r := range path {
		if r == room {
			return true
		}
	}
	return false
}
