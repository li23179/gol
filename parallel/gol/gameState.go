package gol

import (
	"uk.ac.bris.cs/gameoflife/util"
)

type GameState struct {
	World      [][]byte
	AliveCells []util.Cell
	Turn       int
	Pause      bool
}

func (s *GameState) Update(world [][]byte, aliveCells []util.Cell, turn int) {
	s.World = world
	s.AliveCells = aliveCells
	s.Turn = turn
}
