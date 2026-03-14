package dependencies

func AssignAntsToPaths(graph *Graph, paths []*Path) {
	if graph == nil || len(paths) == 0 || graph.Ants <= 0 {
		return
	}

	remainingAnts := graph.Ants

	for remainingAnts > 0 {
		best := paths[0]

		for _, path := range paths {
			if pathScore(path) < pathScore(best) {
				best = path
			}
		}

		best.AntsAssigned++
		remainingAnts--
	}
}

func pathScore(path *Path) int {
	return len(path.Rooms) + path.AntsAssigned
}
