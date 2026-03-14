package dependencies

func SimulateAntMovements(graph *Graph, paths []*Path) [][]Move {
	if graph == nil || len(paths) == 0 || graph.Ants <= 0 {
		return nil
	}

	ants := []*Ant{}
	nextAntID := 1
	finished := 0
	var allTurns [][]Move

	for finished < graph.Ants {
		var turnMoves []Move

		turnMoves = append(turnMoves, moveExistingAnts(ants, &finished)...)
		turnMoves = append(turnMoves, launchNewAnts(paths, &ants, &nextAntID, graph.Ants)...)

		if len(turnMoves) == 0 {
			break
		}

		allTurns = append(allTurns, turnMoves)
	}

	return allTurns
}

func moveExistingAnts(ants []*Ant, finished *int) []Move {
	var moves []Move

	for _, ant := range ants {
		if ant.Finished {
			continue
		}

		nextPos := ant.Position + 1
		if nextPos >= len(ant.Path.Rooms) {
			continue
		}

		room := ant.Path.Rooms[nextPos]
		ant.Position++

		moves = append(moves, Move{
			AntID:    ant.ID,
			RoomName: room.Name,
		})

		if ant.Position == len(ant.Path.Rooms)-1 {
			ant.Finished = true
			*finished += 1
		}
	}

	return moves
}

func launchNewAnts(paths []*Path, ants *[]*Ant, nextAntID *int, totalAnts int) []Move {
	var moves []Move

	for _, path := range paths {
		if path.AntsAssigned <= 0 {
			continue
		}

		if *nextAntID > totalAnts {
			break
		}

		if len(path.Rooms) < 2 {
			continue
		}

		ant := &Ant{
			ID:       *nextAntID,
			Path:     path,
			Position: 1,
			Finished: false,
		}

		*ants = append(*ants, ant)

		moves = append(moves, Move{
			AntID:    ant.ID,
			RoomName: path.Rooms[1].Name,
		})

		path.AntsAssigned--
		*nextAntID += 1
	}

	return moves
}