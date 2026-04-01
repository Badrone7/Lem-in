# 🐜 Lem-in — The Ant Colony Pathfinder

**Lem-in** is a high-performance pathfinding simulation. The goal is to move a specific number of ants from a `##start` room to a `##end` room through a network of connected rooms in the **fewest number of turns** possible.

### 📐 The Challenge
- Each room can only hold **one ant at a time** (except for start and end).
- Tunnels are bidirectional.
- We must find the optimal set of **node-disjoint paths** (paths that don't share any rooms) to maximize ant flow and minimize total turns.

---

## 🚀 How to Run

No build step is required. You can run the program directly using Go:

```bash
# Using go run with the current directory
go run . <input_file>

# OR using go run with main.go
go run main.go <input_file>
```

**Example:**
```bash
go run . test.txt
```

The program will output the original data followed by the results in the format: `L[ant_id]-[room_name]`. Each line represents one turn.

---

## 🧠 Core Algorithm: Edmonds-Karp & Node Splitting

To solve the traffic jam problem efficiently, this project implements the **Edmonds-Karp algorithm** (a specific implementation of the Ford-Fulkerson method) combined with a technique called **Node-Splitting**.

### 1. Node-Splitting (Disjoint Paths)
Standard pathfinding finds the shortest path, but it doesn't prevent multiple paths from "crashing" into the same room. We solve this by splitting every room into two nodes:
- **`Room_In`**: Where flow enters.
- **`Room_Out`**: Where flow leaves.
- An internal edge connects them with a **capacity of 1**. This mathematically guarantees that only one ant path can pass through that room at a time.

### 2. Edmonds-Karp (Max Flow)
The algorithm finds "augmenting paths" in a residual graph:
1. **BFS (Breadth-First Search)**: Finds the shortest available path from start to end.
2. **Push Flow**: Once a path is found, we reduce its capacity to "occupy" it.
3. **Reverse Edges**: We add "undo" edges that allow the algorithm to backtrack and find a better combination of paths if a new, shorter configuration exists.

### 3. Turning Flow into Time
We don't just look for the most paths; we look for the **optimal** number of paths. 
- 1 short path might be better for 5 ants.
- 5 longer paths might be better for 500 ants.
The code calculates the total turns for every new path found and keeps the best-performing set.

---

## 📂 Project Architecture

| File | Responsibility |
| :--- | :--- |
| **`main.go`** | The conductor. It links the parser, pathfinder, and simulator together. |
| **`parser.go`** | Reads the input file, validates the format, and generates the graph. |
| **`pathfinder.go`** | The "Brain." Implements Edmonds-Karp, node-splitting, and path selection. |
| **`simulator.go`** | The "Engine." Handles the turn-by-turn movement of ants along paths. |
| **`assigner.go`** | Logic for distributing ants across found paths based on path length. |
| **`structs.go`** | Definitions for `Room`, `Graph`, `Path`, `Ant`, and `Move`. |

---

## 👤 Author

**bguitoni**
