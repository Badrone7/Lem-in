package dependencies

import "errors"

func FindPaths(graph *Graph) ([]*Path, error) {
	if graph == nil {
		return nil, errors.New("graph is nil")
	}
	if graph.Start == nil || graph.End == nil {
		return nil, errors.New("start or end room not defined")
	}

	candidates, err := findCandidatePaths(graph)
	if err != nil {
		return nil, err
	}

	best := selectBestDisjointPaths(candidates)
	if len(best) == 0 {
		return nil, errors.New("no valid disjoint paths found")
	}

	return best, nil
}

func findCandidatePaths(graph *Graph) ([]*Path, error) {
	queue := [][]*Room{{graph.Start}}
	var found [][]*Room
	minLen := -1
	maxExtra := 2

	for len(queue) > 0 {
		currentPath := queue[0]
		queue = queue[1:]

		lastRoom := currentPath[len(currentPath)-1]

		if minLen != -1 && len(currentPath) > minLen+maxExtra {
			continue
		}

		if lastRoom == graph.End {
			if minLen == -1 {
				minLen = len(currentPath)
			}
			found = append(found, currentPath)
			continue
		}

		for _, neighbor := range lastRoom.Links {
			if containsRoom(currentPath, neighbor) {
				continue
			}

			newPath := append([]*Room{}, currentPath...)
			newPath = append(newPath, neighbor)
			queue = append(queue, newPath)
		}
	}

	if len(found) == 0 {
		return nil, errors.New("no path found")
	}

	paths := make([]*Path, 0, len(found))
	for _, rooms := range found {
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

func selectBestDisjointPaths(paths []*Path) []*Path {
	var best []*Path

	var backtrack func(index int, current []*Path)
	backtrack = func(index int, current []*Path) {
		if index == len(paths) {
			if isBetterPathSet(current, best) {
				best = append([]*Path{}, current...)
			}
			return
		}

		backtrack(index+1, current)

		if canAddPath(current, paths[index]) {
			current = append(current, paths[index])
			backtrack(index+1, current)
		}
	}

	backtrack(0, []*Path{})
	return best
}

func canAddPath(current []*Path, candidate *Path) bool {
	for _, p := range current {
		if pathsConflict(p, candidate) {
			return false
		}
	}
	return true
}

func pathsConflict(a, b *Path) bool {
	used := make(map[*Room]bool)

	for i := 1; i < len(a.Rooms)-1; i++ {
		used[a.Rooms[i]] = true
	}

	for i := 1; i < len(b.Rooms)-1; i++ {
		if used[b.Rooms[i]] {
			return true
		}
	}

	return false
}

func isBetterPathSet(current, best []*Path) bool {
	if len(current) > len(best) {
		return true
	}
	if len(current) < len(best) {
		return false
	}

	return totalPathLength(current) < totalPathLength(best)
}

func totalPathLength(paths []*Path) int {
	total := 0
	for _, p := range paths {
		total += len(p.Rooms)
	}
	return total
}
