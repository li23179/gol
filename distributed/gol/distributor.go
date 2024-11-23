package gol

import (
	"fmt"
	"net/rpc"
	"time"
	"uk.ac.bris.cs/gameoflife/stubs"
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

func loadWorld(p stubs.Params, c distributorChannels) [][]byte {
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

func exportWorld(p stubs.Params, c distributorChannels, finishWorld [][]byte, turn int) {
	checkIoIdle(c)
	outFileName := fmt.Sprintf("%vx%vx%v", p.ImageWidth, p.ImageHeight, turn)
	c.ioCommand <- ioOutput
	c.ioFilename <- outFileName
	for _, w := range finishWorld {
		for i := range w {
			c.ioOutput <- w[i]
		}
	}
	// Make sure all byte in the world has been passed to the output channel before sending ImageOutputComplete
	checkIoIdle(c)
	c.events <- ImageOutputComplete{turn, outFileName}
}

func ManageKeyPress(c distributorChannels, p stubs.Params, client *rpc.Client) {
	keyReq := stubs.KeyPressRequest{}
	for {
		res := new(stubs.GameResponse)
		select {
		case k := <-c.keyPressCh:
			switch k {
			case 's':
				_ = client.Call(stubs.SaveWorld, keyReq, res)
				outFileName := fmt.Sprintf("%vx%vx%v", p.ImageWidth, p.ImageHeight, res.Turn)
				exportWorld(p, c, res.World, res.Turn)
				c.events <- ImageOutputComplete{CompletedTurns: res.Turn, Filename: outFileName}
			case 'q':
				_ = client.Call(stubs.ClientQuit, keyReq, res)
			case 'k':
				_ = client.Call(stubs.ShutDownService, keyReq, res)
				return
			case 'p':
				_ = client.Call(stubs.PauseGame, keyReq, res)
				if res.Paused {
					c.events <- StateChange{CompletedTurns: res.Turn, NewState: Paused}
				} else {
					c.events <- StateChange{CompletedTurns: res.Turn, NewState: Executing}
				}
				fmt.Print(res.Message)
			}
		}
	}
}

func ReportAliveCell(c distributorChannels, client *rpc.Client, quitCh chan bool) {
	ticker := time.NewTicker(2 * time.Second)
	req := stubs.TickerRequest{}
	res := new(stubs.TickerResponse)
	for {
		select {
		case <-ticker.C:
			_ = client.Call(stubs.Ticker, req, res)
			c.events <- AliveCellsCount{
				CompletedTurns: res.Turn,
				CellsCount:     len(res.AliveCells),
			}
		case <-quitCh:
			ticker.Stop()
			return
		}
	}
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p stubs.Params, c distributorChannels) {
	// TODO: Create a 2D slice to store the world.
	//sAddr := flag.String("server", "127.0.0.1:8030", "ip:port string to connect as server")
	//flag.Parse()
	//brokerAddr := flag.String("broker", "127.0.0.1:8030", "Address of Broker instance")
	//flag.Parse()

	client, err := rpc.Dial("tcp", "127.0.0.1:8030")
	util.Check(err)

	inputWorld := loadWorld(p, c)
	turn := 0
	req := stubs.GameRequest{World: inputWorld, P: p, Turn: turn}
	res := new(stubs.GameResponse)
	quitAliveCells := make(chan bool)
	// Client starts gol
	go ManageKeyPress(c, p, client)
	go ReportAliveCell(c, client, quitAliveCells)
	runGol := client.Go(stubs.RunGol, req, res, nil)
	<-runGol.Done
	quitAliveCells <- true

	util.Check(err)
	if res.Kill {
		closeReq := stubs.CloseRequest{}
		closeRes := new(stubs.CloseResponse)
		_ = client.Call(stubs.CloseBroker, closeReq, closeRes)
	}

	c.events <- TurnComplete{CompletedTurns: res.Turn}
	c.events <- FinalTurnComplete{CompletedTurns: res.Turn, Alive: res.AliveCells}

	// output pgm
	exportWorld(p, c, res.World, res.Turn)
	c.events <- StateChange{res.Turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)

	err = client.Close()
	util.Check(err)
}
