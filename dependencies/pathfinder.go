package dependencies

import "errors"

// FindPaths finds the optimal set of node-disjoint paths from Start to End
// using Edmonds-Karp (BFS augmenting paths) with node-splitting.
// It picks the path set that minimizes the number of turns to move all ants.
func FindPaths(graph *Graph) ([]*Path, error) {
	if graph == nil {
		return nil, errors.New("graph is nil")
	}
	if graph.Start == nil || graph.End == nil {
		return nil, errors.New("start or end room not defined")
	}

	rg := buildResidualGraph(graph)

	var bestPaths []*Path
	bestTurns := -1

	for {
		augPath := bfsAugment(rg)
		if augPath == nil {
			break
		}
		applyAugmentation(rg, augPath)

		paths := extractPaths(rg, graph)
		if len(paths) == 0 {
			continue
		}

		turns := computeTotalTurns(graph.Ants, paths)
		if bestTurns == -1 || turns < bestTurns {
			bestTurns = turns
			bestPaths = paths
		}
	}

	if len(bestPaths) == 0 {
		return nil, errors.New("no valid disjoint paths found")
	}
	return bestPaths, nil
}

// --- Residual graph (adjacency-list with capacities) ---
//
// Node-splitting: room index i → nodeIn = 2*i, nodeOut = 2*i+1
// Internal edge: nodeIn → nodeOut (cap 1, or ∞ for start/end)
// External edge per link (u,v): u_out → v_in (cap 1), v_out → u_in (cap 1)

type edge struct {
	to  int
	cap int
	rev int // index of reverse edge in adj[to]
}

type residualGraph struct {
	adj      [][]edge
	n        int
	startOut int
	endIn    int
	idxToRoom []string // roomIndex → room name
}

func buildResidualGraph(graph *Graph) *residualGraph {
	roomToIdx := make(map[string]int, len(graph.Rooms))
	idxToRoom := make([]string, len(graph.Rooms))
	idx := 0
	for name := range graph.Rooms {
		roomToIdx[name] = idx
		idxToRoom[idx] = name
		idx++
	}

	numRooms := len(graph.Rooms)
	totalNodes := numRooms * 2

	rg := &residualGraph{
		adj:       make([][]edge, totalNodes),
		n:         totalNodes,
		startOut:  roomToIdx[graph.Start.Name]*2 + 1,
		endIn:     roomToIdx[graph.End.Name] * 2,
		idxToRoom: idxToRoom,
	}

	addEdge := func(from, to, cap int) {
		rg.adj[from] = append(rg.adj[from], edge{to: to, cap: cap, rev: len(rg.adj[to])})
		rg.adj[to] = append(rg.adj[to], edge{to: from, cap: 0, rev: len(rg.adj[from]) - 1})
	}

	// Internal edges: in → out
	for name, i := range roomToIdx {
		inNode := i * 2
		outNode := i*2 + 1
		if name == graph.Start.Name || name == graph.End.Name {
			addEdge(inNode, outNode, numRooms) // effectively infinite
		} else {
			addEdge(inNode, outNode, 1)
		}
	}

	// External edges: for each link u-v, add u_out→v_in and v_out→u_in
	seen := make(map[[2]int]bool, len(graph.Rooms)*2)
	for _, room := range graph.Rooms {
		u := roomToIdx[room.Name]
		for _, neighbor := range room.Links {
			v := roomToIdx[neighbor.Name]
			key := [2]int{u, v}
			if u > v {
				key = [2]int{v, u}
			}
			if seen[key] {
				continue
			}
			seen[key] = true
			addEdge(u*2+1, v*2, 1)   // u_out → v_in
			addEdge(v*2+1, u*2, 1)   // v_out → u_in
		}
	}

	return rg
}

// bfsAugment finds shortest augmenting path from startOut to endIn.
func bfsAugment(rg *residualGraph) []int {
	parent := make([]int, rg.n)
	parentEdge := make([]int, rg.n)
	for i := range parent {
		parent[i] = -1
	}
	parent[rg.startOut] = rg.startOut

	queue := make([]int, 0, rg.n/4)
	queue = append(queue, rg.startOut)

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		if cur == rg.endIn {
			// Reconstruct path as list of node IDs.
			path := []int{rg.endIn}
			for path[len(path)-1] != rg.startOut {
				path = append(path, parent[path[len(path)-1]])
			}
			for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
				path[i], path[j] = path[j], path[i]
			}
			return path
		}

		for ei, e := range rg.adj[cur] {
			if e.cap <= 0 || parent[e.to] != -1 {
				continue
			}
			parent[e.to] = cur
			parentEdge[e.to] = ei
			queue = append(queue, e.to)
		}
	}
	return nil
}

// applyAugmentation pushes 1 unit of flow along the path.
func applyAugmentation(rg *residualGraph, path []int) {
	for i := 0; i < len(path)-1; i++ {
		u, v := path[i], path[i+1]
		// Find the forward edge u→v and decrease capacity.
		for ei := range rg.adj[u] {
			if rg.adj[u][ei].to == v && rg.adj[u][ei].cap > 0 {
				rg.adj[u][ei].cap--
				rg.adj[v][rg.adj[u][ei].rev].cap++
				break
			}
		}
	}
}

type flowEdge struct {
	to   int
	flow int
}

// extractPaths reconstructs node-disjoint paths from flow state.
func extractPaths(rg *residualGraph, graph *Graph) []*Path {
	// For each forward edge, flow = reverse_edge.cap.

	flowAdj := make([][]flowEdge, rg.n)
	for u := 0; u < rg.n; u++ {
		for _, e := range rg.adj[u] {
			// The reverse edge in adj[e.to][e.rev] tells us flow.
			// Actually, simpler: for each edge pair, the reverse edge's cap = flow.
			// Our forward edges originally had cap > 0. The reverse had cap = 0.
			// After augmentation, reverse.cap = flow amount.
			// So if we look at the reverse entries (those that originally had cap=0),
			// their current cap tells us how much flow is on the forward direction.

			// We want: if this edge (u→e.to) carries flow, add it to flowAdj.
			// Flow on edge u→e.to = cap of reverse edge e.to→u at index e.rev.
			revEdge := rg.adj[e.to][e.rev]
			if revEdge.cap > 0 {
				// This edge carries flow=revEdge.cap, but only if this is a "forward" edge.
				// Forward edges: were added with addEdge(from, to, cap>0).
				// We need to avoid double-counting. The simplest check:
				// An edge is forward if the REVERSE edge originally had cap=0,
				// meaning its current cap = flow > 0 only because of augmentation.
				// The reverse edge's own reverse (which is our edge) has
				// rg.adj[u][?].rev == index in adj[e.to] == e.rev.
				// Actually the simplest approach: only consider edges where flow > 0
				// AND skip the internal start/end high-cap edges if they have no net flow.

				// Simple approach: record directed flow for all edges with positive flow.
				flowAdj[u] = append(flowAdj[u], flowEdge{to: e.to, flow: revEdge.cap})
			}
		}
	}

	// Walk paths from startOut to endIn.
	var paths []*Path

	for {
		path := walkFlow(flowAdj, rg.startOut, rg.endIn, rg.n)
		if path == nil {
			break
		}
		// Convert split-node path to rooms.
		rooms := splitPathToRooms(path, rg, graph)
		if len(rooms) >= 2 {
			paths = append(paths, &Path{Rooms: rooms, AntsAssigned: 0})
		}
	}
	return paths
}

// walkFlow follows flow edges from src to dst, consuming one unit of flow.
func walkFlow(flowAdj [][]flowEdge, src, dst, n int) []int {
	visited := make([]bool, n)
	visited[src] = true
	stack := []int{src}

	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		if cur == dst {
			// Consume flow along this path.
			for i := 0; i < len(stack)-1; i++ {
				u, v := stack[i], stack[i+1]
				for j := range flowAdj[u] {
					if flowAdj[u][j].to == v && flowAdj[u][j].flow > 0 {
						flowAdj[u][j].flow--
						break
					}
				}
			}
			return stack
		}

		found := false
		for i := range flowAdj[cur] {
			if flowAdj[cur][i].flow > 0 && !visited[flowAdj[cur][i].to] {
				visited[flowAdj[cur][i].to] = true
				stack = append(stack, flowAdj[cur][i].to)
				found = true
				break
			}
		}
		if !found {
			stack = stack[:len(stack)-1]
		}
	}
	return nil
}

// splitPathToRooms converts a split-node path to a slice of *Room.
func splitPathToRooms(nodePath []int, rg *residualGraph, graph *Graph) []*Room {
	var rooms []*Room
	prevIdx := -1
	for _, nodeID := range nodePath {
		roomIdx := nodeID / 2
		if roomIdx == prevIdx {
			continue
		}
		prevIdx = roomIdx
		name := rg.idxToRoom[roomIdx]
		if room, ok := graph.Rooms[name]; ok {
			rooms = append(rooms, room)
		}
	}
	return rooms
}

// computeTotalTurns calculates how many turns to move all ants using greedy assignment.
func computeTotalTurns(totalAnts int, paths []*Path) int {
	if len(paths) == 0 || totalAnts == 0 {
		return 0
	}

	lengths := make([]int, len(paths))
	for i, p := range paths {
		lengths[i] = len(p.Rooms) - 1
	}

	assigned := make([]int, len(paths))
	for a := 0; a < totalAnts; a++ {
		bestIdx := 0
		bestCost := lengths[0] + assigned[0]
		for i := 1; i < len(paths); i++ {
			cost := lengths[i] + assigned[i]
			if cost < bestCost {
				bestCost = cost
				bestIdx = i
			}
		}
		assigned[bestIdx]++
	}

	maxTurns := 0
	for i := range paths {
		t := lengths[i] + assigned[i]
		if t > maxTurns {
			maxTurns = t
		}
	}
	return maxTurns
}
