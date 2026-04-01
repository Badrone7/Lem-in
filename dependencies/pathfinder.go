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

	flowNetwork := buildResidualGraph(graph)

	var bestPaths []*Path
	bestTurns := -1

	for {
		augmentingPath := bfsAugment(flowNetwork)
		if augmentingPath == nil {
			break
		}
		applyAugmentation(flowNetwork, augmentingPath)

		currentPaths := extractPaths(flowNetwork, graph)
		if len(currentPaths) == 0 {
			continue
		}

		turns := computeTotalTurns(graph.Ants, currentPaths)
		if bestTurns == -1 || turns < bestTurns {
			bestTurns = turns
			bestPaths = currentPaths
		}
	}

	if len(bestPaths) == 0 {
		return nil, errors.New("no valid disjoint paths found")
	}
	return bestPaths, nil
}

// --- Residual graph (adjacency-list with capacities) ---
//
// Node-splitting: room index i → inNode = 2*i, outNode = 2*i+1
// Internal edge: inNode → outNode (capacity 1, or ∞ for start/end)
// External edge per link (fromRoom, toRoom): fromRoom_out → toRoom_in (cap 1), toRoom_out → fromRoom_in (cap 1)

type edge struct {
	toNode       int
	capacity     int
	reverseIndex int
	isReverse    bool
}

type residualGraph struct {
	adjacency   [][]edge
	nodeCount   int
	startOut    int
	endIn       int
	indexToRoom []string // roomIndex → room name
}

func buildResidualGraph(graph *Graph) *residualGraph {
	roomToIndex := make(map[string]int, len(graph.Rooms))
	indexToRoom := make([]string, len(graph.Rooms))
	counter := 0
	for roomName := range graph.Rooms {
		roomToIndex[roomName] = counter
		indexToRoom[counter] = roomName
		counter++
	}

	numRooms := len(graph.Rooms)
	totalNodes := numRooms * 2

	flowNetwork := &residualGraph{
		adjacency:   make([][]edge, totalNodes),
		nodeCount:   totalNodes,
		startOut:    roomToIndex[graph.Start.Name]*2 + 1,
		endIn:       roomToIndex[graph.End.Name] * 2,
		indexToRoom: indexToRoom,
	}

	addEdge := func(fromNode, toNode, edgeCapacity int) {
		flowNetwork.adjacency[fromNode] = append(flowNetwork.adjacency[fromNode], edge{
			toNode:       toNode,
			capacity:     edgeCapacity,
			reverseIndex: len(flowNetwork.adjacency[toNode]),
			isReverse:    false,
		})
		flowNetwork.adjacency[toNode] = append(flowNetwork.adjacency[toNode], edge{
			toNode:       fromNode,
			capacity:     0,
			reverseIndex: len(flowNetwork.adjacency[fromNode]) - 1,
			isReverse:    true,
		})
	}

	// Internal edges: inNode → outNode (splits each room so at most 1 ant passes through)
	for roomName, roomIndex := range roomToIndex {
		inNode := roomIndex * 2
		outNode := roomIndex*2 + 1
		if roomName == graph.Start.Name || roomName == graph.End.Name {
			addEdge(inNode, outNode, numRooms) // effectively infinite capacity for start/end
		} else {
			addEdge(inNode, outNode, 1) // only 1 ant allowed through this room at a time
		}
	}

	// External edges: for each tunnel (fromRoom ↔ toRoom), add directed edges both ways
	alreadySeen := make(map[[2]int]bool, len(graph.Rooms)*2)
	for _, room := range graph.Rooms {
		fromRoomIndex := roomToIndex[room.Name]
		for _, neighbor := range room.Links {
			toRoomIndex := roomToIndex[neighbor.Name]
			edgePair := [2]int{fromRoomIndex, toRoomIndex}
			if fromRoomIndex > toRoomIndex {
				edgePair = [2]int{toRoomIndex, fromRoomIndex}
			}
			if alreadySeen[edgePair] {
				continue
			}
			alreadySeen[edgePair] = true
			addEdge(fromRoomIndex*2+1, toRoomIndex*2, 1) // fromRoom_out → toRoom_in
			addEdge(toRoomIndex*2+1, fromRoomIndex*2, 1) // toRoom_out → fromRoom_in
		}
	}

	return flowNetwork
}

// bfsAugment finds the shortest augmenting path from startOut to endIn using BFS.
func bfsAugment(flowNetwork *residualGraph) []int {
	parentNode := make([]int, flowNetwork.nodeCount)
	parentEdgeIndex := make([]int, flowNetwork.nodeCount)
	for nodeIndex := range parentNode {
		parentNode[nodeIndex] = -1
	}
	parentNode[flowNetwork.startOut] = flowNetwork.startOut

	queue := make([]int, 0, flowNetwork.nodeCount/4)
	queue = append(queue, flowNetwork.startOut)

	for len(queue) > 0 {
		currentNode := queue[0]
		queue = queue[1:]

		if currentNode == flowNetwork.endIn {
			augmentingPath := []int{flowNetwork.endIn}
			for augmentingPath[len(augmentingPath)-1] != flowNetwork.startOut {
				augmentingPath = append(augmentingPath, parentNode[augmentingPath[len(augmentingPath)-1]])
			}
			for left, right := 0, len(augmentingPath)-1; left < right; left, right = left+1, right-1 {
				augmentingPath[left], augmentingPath[right] = augmentingPath[right], augmentingPath[left]
			}
			return augmentingPath
		}

		for edgeIndex, neighborEdge := range flowNetwork.adjacency[currentNode] {
			if neighborEdge.capacity <= 0 || parentNode[neighborEdge.toNode] != -1 {
				continue
			}
			parentNode[neighborEdge.toNode] = currentNode
			parentEdgeIndex[neighborEdge.toNode] = edgeIndex
			queue = append(queue, neighborEdge.toNode)
		}
	}
	return nil
}

// applyAugmentation pushes 1 unit of flow along the augmenting path.
func applyAugmentation(flowNetwork *residualGraph, augmentingPath []int) {
	for stepIndex := 0; stepIndex < len(augmentingPath)-1; stepIndex++ {
		fromNode, toNode := augmentingPath[stepIndex], augmentingPath[stepIndex+1]
		for edgeIndex := range flowNetwork.adjacency[fromNode] {
			if flowNetwork.adjacency[fromNode][edgeIndex].toNode == toNode &&
				flowNetwork.adjacency[fromNode][edgeIndex].capacity > 0 {
				flowNetwork.adjacency[fromNode][edgeIndex].capacity--
				flowNetwork.adjacency[toNode][flowNetwork.adjacency[fromNode][edgeIndex].reverseIndex].capacity++
				break
			}
		}
	}
}

type flowEdge struct {
	toNode     int
	flowAmount int
}

// extractPaths reconstructs node-disjoint paths from the current flow state.
func extractPaths(flowNetwork *residualGraph, graph *Graph) []*Path {
	flowAdjacency := make([][]flowEdge, flowNetwork.nodeCount)
	for fromNode := 0; fromNode < flowNetwork.nodeCount; fromNode++ {
		for _, currentEdge := range flowNetwork.adjacency[fromNode] {
			if currentEdge.isReverse {
				continue
			}
			reverseEdge := flowNetwork.adjacency[currentEdge.toNode][currentEdge.reverseIndex]
			if reverseEdge.capacity > 0 {
				flowAdjacency[fromNode] = append(flowAdjacency[fromNode], flowEdge{
					toNode:     currentEdge.toNode,
					flowAmount: reverseEdge.capacity,
				})
			}
		}
	}

	var paths []*Path

	for {
		nodePath := walkFlow(flowAdjacency, flowNetwork.startOut, flowNetwork.endIn, flowNetwork.nodeCount)
		if nodePath == nil {
			break
		}
		rooms := splitPathToRooms(nodePath, flowNetwork, graph)
		if len(rooms) >= 2 {
			paths = append(paths, &Path{Rooms: rooms, AntsAssigned: 0})
		}
	}
	return paths
}

// walkFlow follows flow edges from source to destination, consuming one unit of flow per call.
func walkFlow(flowAdjacency [][]flowEdge, source, destination, nodeCount int) []int {
	visited := make([]bool, nodeCount)
	visited[source] = true
	stack := []int{source}

	for len(stack) > 0 {
		currentNode := stack[len(stack)-1]
		if currentNode == destination {
			// Consume 1 unit of flow along the path we just walked.
			for stepIndex := 0; stepIndex < len(stack)-1; stepIndex++ {
				fromNode, toNode := stack[stepIndex], stack[stepIndex+1]
				for flowEdgeIndex := range flowAdjacency[fromNode] {
					if flowAdjacency[fromNode][flowEdgeIndex].toNode == toNode &&
						flowAdjacency[fromNode][flowEdgeIndex].flowAmount > 0 {
						flowAdjacency[fromNode][flowEdgeIndex].flowAmount--
						break
					}
				}
			}
			return stack
		}

		foundNextNode := false
		for neighborIndex := range flowAdjacency[currentNode] {
			neighbor := flowAdjacency[currentNode][neighborIndex]
			if neighbor.flowAmount > 0 && !visited[neighbor.toNode] {
				visited[neighbor.toNode] = true
				stack = append(stack, neighbor.toNode)
				foundNextNode = true
				break
			}
		}
		if !foundNextNode {
			stack = stack[:len(stack)-1] // backtrack
		}
	}
	return nil
}

// splitPathToRooms converts a split-node path (inNode/outNode pairs) to a slice of *Room.
func splitPathToRooms(nodePath []int, flowNetwork *residualGraph, graph *Graph) []*Room {
	var rooms []*Room
	previousRoomIndex := -1
	for _, nodeID := range nodePath {
		roomIndex := nodeID / 2
		if roomIndex == previousRoomIndex {
			continue
		}
		previousRoomIndex = roomIndex
		roomName := flowNetwork.indexToRoom[roomIndex]
		if room, ok := graph.Rooms[roomName]; ok {
			rooms = append(rooms, room)
		}
	}
	return rooms
}

// computeTotalTurns calculates how many turns are needed to move all ants
// by greedily assigning each ant to whichever path finishes earliest.
func computeTotalTurns(totalAnts int, paths []*Path) int {
	if len(paths) == 0 || totalAnts == 0 {
		return 0
	}

	pathLengths := make([]int, len(paths))
	for pathIndex, currentPath := range paths {
		pathLengths[pathIndex] = len(currentPath.Rooms) - 1
	}

	antsOnPath := make([]int, len(paths))
	for antIndex := 0; antIndex < totalAnts; antIndex++ {
		bestPathIndex := 0
		lowestCost := pathLengths[0] + antsOnPath[0]
		for pathIndex := 1; pathIndex < len(paths); pathIndex++ {
			pathCost := pathLengths[pathIndex] + antsOnPath[pathIndex]
			if pathCost < lowestCost {
				lowestCost = pathCost
				bestPathIndex = pathIndex
			}
		}
		antsOnPath[bestPathIndex]++
	}

	maxTurns := 0
	for pathIndex := range paths {
		pathTotalTurns := pathLengths[pathIndex] + antsOnPath[pathIndex]
		if pathTotalTurns > maxTurns {
			maxTurns = pathTotalTurns
		}
	}
	return maxTurns
}
