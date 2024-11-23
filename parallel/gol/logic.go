package gol

import (
	"uk.ac.bris.cs/gameoflife/util"
)

const dead byte = 0
const live byte = 255

func CalculateLiveNeighbour(p Params, x, y int, immutableWorld func(int, int) byte) int {
	counter := 0
	relativeDir := [8][2]int{
		{0, 1}, {0, -1},
		{1, 0}, {-1, 0},
		{-1, -1}, {-1, 1},
		{1, 1}, {1, -1},
	}

	for _, dir := range relativeDir {
		ny := (dir[1] + p.ImageHeight + y) % p.ImageHeight
		nx := (dir[0] + p.ImageWidth + x) % p.ImageWidth

		if immutableWorld(ny, nx) == live {
			counter++
		}
	}
	return counter
}

func CalculateNextState(p Params, startY, endY int, immutableWorld func(int, int) byte) [][]byte {
	// TODO : Implement a parallel version for workers
	newWorld := util.MakeWorld(p.ImageWidth, endY-startY)
	//fmt.Printf("%v \n", startY)
	for y := startY; y < endY; y++ {
		j := y - startY // adjust the column
		for x := 0; x < p.ImageWidth; x++ {
			counter := CalculateLiveNeighbour(p, x, y, immutableWorld)
			if immutableWorld(y, x) == live {
				if counter < 2 || counter > 3 {
					newWorld[j][x] = dead
				} else {
					newWorld[j][x] = live
				}
			} else {
				if counter == 3 {
					newWorld[j][x] = live
				}
			}
		}
		//fmt.Printf("For loop y: %v \n", y)
	}
	return newWorld
}

func CalculateAliveCells(p Params, startY, endY int, immutableWorld func(int, int) byte) []util.Cell {
	var cells []util.Cell
	for y := startY; y < endY; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			if immutableWorld(y, x) == live {
				cells = append(cells, util.Cell{X: x, Y: y})
			}
		}
	}
	return cells
}

// StateWorker worldWorker work on the same board
// use goroutine to distribute the worker on different section and export via channel
func StateWorker(p Params, startY, endY int, immutableWorld func(int, int) byte, worldCh chan<- [][]byte) {
	partialWorld := CalculateNextState(p, startY, endY, immutableWorld)
	worldCh <- partialWorld
	//fmt.Println("finish worldCh")
}

func AliveCellWorker(p Params, startY, endY int, immutableWorld func(int, int) byte, aliveCellCh chan<- []util.Cell) {
	partialAliveCells := CalculateAliveCells(p, startY, endY, immutableWorld)
	aliveCellCh <- partialAliveCells
}

// DelegateStateWork accumulate workload for each worker for each turn
func DelegateStateWork(p Params, immutableWorld func(int, int) byte, worldChs []chan [][]byte, finishWorldCh chan<- [][]byte) {
	baseWorkload := p.ImageHeight / p.Threads
	extraWorkerThreads := p.ImageHeight % p.Threads

	startY := 0
	var finishWorld [][]byte

	for t := 0; t < p.Threads; t++ {
		workload := baseWorkload
		if t < extraWorkerThreads {
			workload++
		}
		endY := startY + workload
		go StateWorker(p, startY, endY, immutableWorld, worldChs[t])
		startY = endY
	}

	for t := 0; t < p.Threads; t++ {
		finishWorld = append(finishWorld, <-worldChs[t]...)
	}
	finishWorldCh <- finishWorld
}

func DelegateCellWork(p Params, immutableWorld func(int, int) byte, aliveCellChs []chan []util.Cell, finishCellsCh chan<- []util.Cell) {
	baseWorkload := p.ImageHeight / p.Threads
	extraWorkerThreads := p.ImageHeight % p.Threads

	startY := 0

	var finishCells []util.Cell

	for t := 0; t < p.Threads; t++ {
		workload := baseWorkload
		if t < extraWorkerThreads {
			workload++
		}
		endY := startY + workload
		go AliveCellWorker(p, startY, endY, immutableWorld, aliveCellChs[t])
		startY = endY
	}

	for t := 0; t < p.Threads; t++ {
		finishCells = append(finishCells, <-aliveCellChs[t]...)
	}
	finishCellsCh <- finishCells
}

func calculateFlippedCells(aliveCells []util.Cell, newAliveCells []util.Cell) []util.Cell {
	// Create maps to store the frequency of each cell in both arrays
	cellMap := make(map[util.Cell]int)

	for _, cell := range aliveCells {
		cellMap[cell]++
	}

	for _, cell := range newAliveCells {
		cellMap[cell]++
	}

	// Result slice to store the symmetric difference
	var flippedCells []util.Cell

	// Add cells that appear only once, meaning they are present in either array but not in both
	for cell, count := range cellMap {
		if count == 1 {
			flippedCells = append(flippedCells, cell)
		}
	}

	return flippedCells
}
