package dependencies

// This file defines the core data structures used in the program.

type Room struct {
	Name  string
	X     int
	Y     int
	Links []*Room
}

type Graph struct {
	Ants  int
	Rooms map[string]*Room
	Start *Room
	End   *Room
}

type Path struct {
	Rooms        []*Room
	AntsAssigned int
}

type Ant struct {
	ID       int
	Path     *Path
	Position int
	Finished bool
}

type Move struct {
	AntID    int
	RoomName string
}