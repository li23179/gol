package gol

import (
	"uk.ac.bris.cs/gameoflife/util"
)

type WorkerChannels struct {
	StateWorkerChannels   []chan [][]byte
	CellWorkerChannels    []chan []util.Cell
	NextStateChannel      chan [][]byte
	NextAliveCellsChannel chan []util.Cell
}

func (w *WorkerChannels) InitialiseChannels(p Params) {
	w.StateWorkerChannels = util.MakeStateWorkerChannels(p.Threads)
	w.CellWorkerChannels = util.MakeCellWorkerChannels(p.Threads)
	w.NextStateChannel = util.MakeNextStateChannel()
	w.NextAliveCellsChannel = util.MakeNextAliveCellChannel()
}

type KeyPressChannels struct {
	SaveChannel  chan bool
	QuitChannel  chan bool
	PauseChannel chan bool
	Done         chan bool
}

func (k *KeyPressChannels) InitialiseChannels() {
	k.SaveChannel = util.MakeBoolChannel()
	k.QuitChannel = util.MakeBoolChannel()
	k.PauseChannel = util.MakeBoolChannel()
}
