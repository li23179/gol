package util

func MakeWorld(x, y int) [][]byte {
	world := make([][]byte, y)
	for i := range world {
		world[i] = make([]byte, x)
	}
	return world
}

// MakeImmutableWorld Use closure, no direct access to world and can modify it
// allow to pass the closure to multiple goroutines without causing any potential race condition
func MakeImmutableWorld(world [][]byte) func(y, x int) byte {
	return func(y, x int) byte {
		return world[y][x]
	}
}

func MakeStateWorkerChannels(threads int) []chan [][]byte {
	chs := make([]chan [][]byte, threads)
	for i := 0; i < threads; i++ {
		chs[i] = make(chan [][]byte)
	}
	return chs
}

func MakeCellWorkerChannels(threads int) []chan []Cell {
	chs := make([]chan []Cell, threads)
	for i := 0; i < threads; i++ {
		chs[i] = make(chan []Cell)
	}
	return chs
}

func MakeNextStateChannel() chan [][]byte {
	return make(chan [][]byte)
}

func MakeNextAliveCellChannel() chan []Cell {
	return make(chan []Cell)
}

func MakeBoolChannel() chan bool {
	return make(chan bool)
}
