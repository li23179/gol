package gol

import (
	"fmt"
	"sync"
	"time"
	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
	keyPressCh <-chan rune
}

func checkIoIdle(c distributorChannels) {
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
}

func loadWorld(p Params, c distributorChannels) [][]byte {
	inFileName := fmt.Sprintf("%vx%v", p.ImageWidth, p.ImageHeight)
	checkIoIdle(c)
	// start reading the pgm file if io is idle
	c.ioCommand <- ioInput
	c.ioFilename <- inFileName
	world := util.MakeWorld(p.ImageWidth, p.ImageHeight)
	for _, w := range world {
		for i := range w {
			w[i] = <-c.ioInput
		}
	}
	return world
}

func exportWorld(p Params, c distributorChannels, state GameState) {
	checkIoIdle(c)
	outFileName := fmt.Sprintf("%vx%vx%v", p.ImageWidth, p.ImageHeight, state.Turn)
	c.ioCommand <- ioOutput
	c.ioFilename <- outFileName
	for _, w := range state.World {
		for i := range w {
			c.ioOutput <- w[i]
		}
	}
	// Make sure all byte in the world has been passed to the output channel before sending ImageOutputComplete
	checkIoIdle(c)
	c.events <- ImageOutputComplete{state.Turn, outFileName}
}

func reportAliveCells(c distributorChannels, mu *sync.Mutex, quitCh <-chan bool) {
	ticker := time.NewTicker(2 * time.Second)
	for {
		select {
		case <-ticker.C:
			mu.Lock()
			c.events <- AliveCellsCount{CompletedTurns: gameState.Turn, CellsCount: len(gameState.AliveCells)}
			mu.Unlock()
		case <-quitCh:
			ticker.Stop()
			return
		}
	}
}

func manageKeyPress(c distributorChannels, chs *KeyPressChannels, quitCh chan bool) {
	for {
		select {
		case key := <-c.keyPressCh:
			switch key {
			case 's':
				chs.SaveChannel <- true
			case 'q':
				chs.QuitChannel <- true
			case 'p':
				chs.PauseChannel <- true
			}
		case <-quitCh:
			break
		}
	}
}

func quitGol(c distributorChannels, p Params, quitAliveCellsCh chan<- bool, quitKeyPress chan<- bool) {
	quitAliveCellsCh <- true
	quitKeyPress <- true

	c.events <- FinalTurnComplete{CompletedTurns: gameState.Turn, Alive: gameState.AliveCells}

	exportWorld(p, c, gameState)

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{gameState.Turn, Quitting}
}

var gameState GameState

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {

	var stateMutex sync.Mutex
	gameState.Pause = false

	// TODO: Create a 2D slice to store the world.
	inputWorld := loadWorld(p, c)
	turn := 0

	immutableWorld := util.MakeImmutableWorld(inputWorld)
	aliveCells := CalculateAliveCells(p, 0, p.ImageHeight, immutableWorld)

	workerChs := new(WorkerChannels)
	workerChs.InitialiseChannels(p)

	keyPressChs := new(KeyPressChannels)
	keyPressChs.InitialiseChannels()

	c.events <- CellsFlipped{turn, aliveCells}
	c.events <- StateChange{turn, Executing}

	gameState.Update(inputWorld, aliveCells, turn)

	quitAliveCellsCh := make(chan bool)
	quitKeyPress := make(chan bool)

	go reportAliveCells(c, &stateMutex, quitAliveCellsCh)
	go manageKeyPress(c, keyPressChs, quitKeyPress)

	// TODO: Execute all turns of the Game of Life.
	for turn < p.Turns {
		// if receive key signal process it, otherwise run gol
		select {
		case <-keyPressChs.SaveChannel:
			stateMutex.Lock()
			exportWorld(p, c, gameState)
			stateMutex.Unlock()

		case <-keyPressChs.PauseChannel:
			stateMutex.Lock()
			gameState.Pause = !gameState.Pause
			if gameState.Pause {
				c.events <- StateChange{CompletedTurns: gameState.Turn, NewState: Paused}
			} else {
				c.events <- StateChange{CompletedTurns: gameState.Turn, NewState: Executing}
			}
			stateMutex.Unlock()

		case <-keyPressChs.QuitChannel:
			// quit all goroutines
			stateMutex.Lock()
			quitGol(c, p, quitAliveCellsCh, quitKeyPress)
			stateMutex.Unlock()
			close(c.events)
			return

		default:
			if !gameState.Pause {
				turn++
				go DelegateStateWork(p, immutableWorld, workerChs.StateWorkerChannels, workerChs.NextStateChannel)
				nextStateWorld := <-workerChs.NextStateChannel

				immutableWorld = util.MakeImmutableWorld(nextStateWorld)

				go DelegateCellWork(p, immutableWorld, workerChs.CellWorkerChannels, workerChs.NextAliveCellsChannel)
				nextAliveCells := <-workerChs.NextAliveCellsChannel

				flipped := calculateFlippedCells(aliveCells, nextAliveCells)

				stateMutex.Lock()
				gameState.Update(nextStateWorld, nextAliveCells, turn)
				c.events <- CellsFlipped{CompletedTurns: gameState.Turn, Cells: flipped}
				c.events <- TurnComplete{CompletedTurns: gameState.Turn}
				stateMutex.Unlock()

				aliveCells = nextAliveCells
			}
		}
	}

	stateMutex.Lock()
	quitGol(c, p, quitAliveCellsCh, quitKeyPress)
	stateMutex.Unlock()

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
