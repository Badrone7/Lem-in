# 🐜 Lem-in — Ant Colony Pathfinder

Lem-in is a program that simulates an ant farm. You have a graph of rooms connected by tunnels, a **start** room, an **end** room, and a bunch of ants. The goal is to move all the ants from start to end in the **fewest number of turns** possible, with one constraint: **only one ant can be in a room at a time** (except start and end).

---

## 📦 How to Build & Run

```bash
# Build
go build -o lem-in .

# Run
./lem-in <input_file>

# Example
./lem-in big_7.txt
```

The output is a series of lines. Each line is one "turn", and looks like:

```
L1-roomA L2-roomB L3-roomC
```

This means: ant 1 moved to `roomA`, ant 2 moved to `roomB`, ant 3 moved to `roomC`, all in the same turn.

---

## 📄 Input File Format

```
<number_of_ants>
<rooms>
<links>
```

**Example:**

```
3
##start
start 0 0
##end
end 4 4
room1 1 1
room2 2 2
room3 3 3
start-room1
room1-room2
room2-room3
room3-end
start-room2
room2-end
```

- First line: number of ants (here: 3)
- `##start` / `##end`: the next line defines the start/end room
- Rooms: `name x y` (name + coordinates)
- Links: `room1-room2` (a tunnel between two rooms)

---

## 🧠 How It Works — The Big Picture

The program runs in **4 steps**:

```
1. PARSE  →  2. FIND PATHS  →  3. ASSIGN ANTS  →  4. SIMULATE
```

### Step 1: Parse (`parser.go`)

Reads the input file and builds a **graph** — a data structure where each room knows which other rooms it connects to.

### Step 2: Find Paths (`pathfinder.go`) ⭐ The Hard Part

Finds the **best set of paths** from start to end. This is where all the magic happens (explained in detail below).

### Step 3: Assign Ants (`assigner.go`)

Distributes ants across the paths. Shorter paths get more ants because they're faster.

### Step 4: Simulate (`simulator.go`)

Moves ants along their assigned paths one step at a time, outputting the moves per turn.

---

## 🔍 The Pathfinder — Explained Step by Step

This is the core algorithm. Let me break it down in simple terms.

### The Problem

We need to find **multiple paths** from start to end that:

1. **Don't share any rooms** (node-disjoint) — because only one ant can be in a room at a time
2. Together, allow all ants to finish in the **fewest turns**

### Why This Is Hard

Imagine you have 100 possible paths. You can't just pick the shortest one — if you have 1000 ants, one path creates a traffic jam. You want *multiple* paths running in parallel. But the paths can't share rooms (except start/end).

Finding the best combination of non-overlapping paths is extremely hard if you try all combinations (2^100 possibilities). The old code did exactly that — which is why it crashed on big inputs.

### The Solution: Max-Flow with Node-Splitting

We use a well-known algorithm from graph theory called **Edmonds-Karp** (a version of the Ford-Fulkerson maximum flow algorithm). Here's how it works, step by step:

---

#### 🧱 Step A: Node-Splitting

**The idea:** We need to make sure each room is used by at most one path. We do this by splitting every room into two "virtual" nodes:

```
Original:         After splitting:

  [room1]    →    [room1_in] ---(capacity 1)--→ [room1_out]
```

- `room1_in` = the "entrance" of the room
- `room1_out` = the "exit" of the room
- The edge between them has **capacity 1** — meaning only **one path** can go through this room

For **start** and **end** rooms, the capacity is unlimited (many ants can leave start / arrive at end).

For tunnel connections, we add edges between the **out** of one room and the **in** of another:

```
If rooms A and B are connected:

  A_out → B_in   (capacity 1)
  B_out → A_in   (capacity 1)
```

This creates a **flow network** — a graph where each edge has a maximum capacity.

---

#### 🌊 Step B: Finding Augmenting Paths (BFS)

Now we iteratively find paths through this network:

1. **Run BFS** (Breadth-First Search) from `start_out` to `end_in`
2. BFS guarantees we find the **shortest** available path first
3. If we find a path, we "push flow" through it (explained next)
4. Repeat until no more paths can be found

**What is BFS?** It's like exploring a maze level by level. First you check all rooms 1 step away, then all rooms 2 steps away, etc. This guarantees you find the shortest path.

---

#### 🔄 Step C: Pushing Flow (The Residual Graph)

When we find an augmenting path, we **flip the edges**:

```
Before augmenting through A→B:
  A ---→ B  (capacity 1)    A ←--- B  (capacity 0)

After augmenting:
  A ---→ B  (capacity 0)    A ←--- B  (capacity 1)
```

- Forward edge: capacity goes from 1 → 0 (used up)
- Reverse edge: capacity goes from 0 → 1 (can "undo" this flow)

**Why reverse edges?** This is the genius of the algorithm. Sometimes, the first path you find isn't optimal. The reverse edges let the algorithm "undo" a previous choice and reroute flow through a better combination. The math guarantees this always finds the maximum number of non-overlapping paths.

---

#### 🏆 Step D: Picking the Best Set

After each new path is found, we:

1. **Extract all current paths** from the flow (by following edges that carry flow)
2. **Calculate how many turns** it would take to move all ants through these paths
3. **Keep the best set** (fewest turns)

The turn calculation works like this:

```
turns = max over all paths of (path_length + ants_on_that_path)
```

More paths = less congestion, but paths might be longer. The algorithm finds the sweet spot.

---

#### 🎯 Why This Is Fast

| | Old Algorithm | New Algorithm |
|---|---|---|
| **Approach** | Enumerate ALL paths, try ALL combinations | Find paths one by one via BFS |
| **Time** | O(2^N) — exponential 💀 | O(V × E) — polynomial ✅ |
| **Memory** | Stores all paths at once 💀 | Only one BFS queue at a time ✅ |
| **23K rooms** | SIGKILL (out of memory) | 4.4 seconds |

Where V = number of rooms, E = number of tunnels.

---

## 📂 Project Structure

```
Lem-in/
├── main.go                    # Entry point — orchestrates all steps
├── dependencies/
│   ├── structs.go             # Data types (Room, Graph, Path, Ant, Move)
│   ├── parser.go              # Step 1: Reads input file → Graph
│   ├── pathfinder.go          # Step 2: Edmonds-Karp max-flow pathfinder
│   ├── assigner.go            # Step 3: Distributes ants across paths
│   └── simulator.go           # Step 4: Simulates ant movement turn by turn
├── big_7.txt                  # Test file (~1K rooms)
├── pylone_400_10_10_35_35_3_no_z.txt  # Test file (~1.3K rooms)
├── test.txt                   # Large test file (~23K rooms, 4K ants)
└── go.mod
```

---

## 📊 Data Structures

### `Room`
```go
type Room struct {
    Name  string   // Room name (e.g., "room1")
    X, Y  int      // Coordinates (for the input format)
    Links []*Room  // Which rooms this room connects to
}
```

### `Graph`
```go
type Graph struct {
    Ants  int              // Number of ants
    Rooms map[string]*Room // All rooms by name
    Start *Room            // The start room
    End   *Room            // The end room
}
```

### `Path`
```go
type Path struct {
    Rooms        []*Room // Ordered list of rooms from start to end
    AntsAssigned int     // How many ants will use this path
}
```

---

## 🧪 Testing

```bash
# Small test
time ./lem-in big_7.txt

# Medium test
time ./lem-in pylone_400_10_10_35_35_3_no_z.txt

# Large test (previously crashed — should now complete in ~5 seconds)
time ./lem-in test.txt
```

---

## 📚 Key Concepts Summary

| Concept | What It Means |
|---|---|
| **Node-disjoint paths** | Paths that don't share any rooms (except start/end) |
| **Max-flow** | Finding the maximum amount of "stuff" you can push through a network |
| **Node-splitting** | Turning a room into two nodes to limit flow through it |
| **BFS** | Exploring a graph level by level to find shortest paths |
| **Residual graph** | The graph after subtracting used flow, with reverse edges added |
| **Augmenting path** | A path from start to end that still has available capacity |
| **Edmonds-Karp** | Max-flow algorithm that uses BFS to find augmenting paths |

---

## 👤 Author

Built as part of the Zone01 curriculum.
