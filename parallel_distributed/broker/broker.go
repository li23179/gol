package main

import (
	"flag"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"sync"
	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

var (
	workers   []Worker
	gameState GameState
	mu        sync.RWMutex // allow multiple Read and one Write
	stateMu   sync.RWMutex
	keyCh     = KeyPressChannels{
		Quit:   make(chan bool),
		Kill:   make(chan bool),
		Paused: make(chan bool),
	}
	quitBroker = make(chan bool)
)

type Worker struct {
	IP string
	*rpc.Client
}

type KeyPressChannels struct {
	Quit   chan bool
	Kill   chan bool
	Paused chan bool
}

type GameState struct {
	World      [][]byte
	AliveCells []util.Cell
	Turn       int
	Workers    int
	Paused     bool
	Resume     bool
}

func (state *GameState) Save(res *stubs.GameResponse) {
	state.World = res.World
	state.AliveCells = res.AliveCells
	state.Turn = res.Turn
}

func (state *GameState) Load(res *stubs.GameResponse) {
	res.World = state.World
	res.AliveCells = state.AliveCells
	res.Turn = state.Turn
}

// split the slice of world
func split(p stubs.Params, startY, endY int, world [][]byte) [][]byte {
	sliceWorld := util.MakeWorld(p.ImageWidth, endY-startY)
	for y := startY; y < endY; y++ {
		j := y - startY // offset for sliceWorld
		for x := 0; x < p.ImageWidth; x++ {
			sliceWorld[j][x] = world[y][x]
		}
	}
	return sliceWorld
}

func calculateWorkload(p stubs.Params) (baseWorkload, extraWorkers int) {
	// calculate the workload for each worker
	mu.Lock()
	nodes := len(workers)
	mu.Unlock()
	baseWorkload = p.ImageHeight / nodes
	extraWorkers = p.ImageHeight % nodes
	return
}

func haloExchange() {
	mu.Lock()
	for _, worker := range workers {
		bReq := stubs.BrokerRequest{}
		bRes := new(stubs.BrokerResponse)
		err := worker.Call(stubs.HaloExchange, bReq, bRes)
		if err != nil {
			fmt.Println("RPC call error:", err)
		}
	}
	mu.Unlock()
}

func calculateStateAndCells(res *stubs.GameResponse, turn int) {
	var aliveCells []util.Cell
	var state [][]byte

	for _, w := range workers {
		pReq := stubs.ProcessRequest{}
		pRes := new(stubs.ProcessResponse)

		err := w.Call("Worker.ProcessTurn", pReq, pRes)
		if err != nil {
			fmt.Println("RPC call error:", err)
		}
		state = append(state, pRes.PartialWorld...)
		aliveCells = append(aliveCells, pRes.PartialAliveCells...)
	}

	res.World = state
	res.AliveCells = aliveCells
	res.Turn = turn
}

func initialise(req stubs.GameRequest, res *stubs.GameResponse, turn int) {
	p := req.P
	world := req.World

	baseWorkload, extraWorkers := calculateWorkload(p)
	startY := 0

	var aliveCells []util.Cell
	var state [][]byte

	mu.Lock()
	for i, worker := range workers {
		workload := baseWorkload
		if i < extraWorkers {
			workload++
		}

		endY := startY + workload
		sliceWorld := split(p, startY, endY, world)

		prevIndex := (i - 1 + len(workers)) % len(workers)
		nextIndex := (i + 1 + len(workers)) % len(workers)

		bReq := stubs.BrokerRequest{
			PartialWorld: sliceWorld,
			P:            p,
			StartY:       startY,
			PrevAddr:     workers[prevIndex].IP,
			NextAddr:     workers[nextIndex].IP,
			Workers:      len(workers),
			IP:           worker.IP,
		}

		fmt.Println(len(workers))
		bRes := new(stubs.BrokerResponse)

		_ = worker.Call(stubs.Initialise, bReq, bRes)

		state = append(state, bRes.PartialWorld...)
		aliveCells = append(aliveCells, bRes.PartialAliveCells...)
		startY = endY
	}
	mu.Unlock()
	res.World = state
	res.AliveCells = aliveCells
	res.Turn = turn
}

func CloseServer() {
	mu.Lock()
	for _, w := range workers {
		req := stubs.CloseRequest{}
		res := new(stubs.CloseResponse)
		_ = w.Call(stubs.CloseServer, req, res)
	}
	mu.Unlock()
}

type Broker struct{}

func (b *Broker) Register(req stubs.WorkerRequest, res *stubs.WorkerResponse) (err error) {
	client, _ := rpc.Dial("tcp", req.IP)
	worker := Worker{
		IP:     req.IP,
		Client: client,
	}
	mu.Lock()
	workers = append(workers, worker)
	mu.Unlock()
	res.Message = "Registered Successfully"
	return
}

func (b *Broker) SaveWorld(req stubs.KeyPressRequest, res *stubs.GameResponse) (err error) {
	stateMu.RLock()
	gameState.Load(res)
	stateMu.RUnlock()
	return
}

func (b *Broker) ShutDownService(req stubs.KeyPressRequest, res *stubs.GameResponse) (err error) {
	stateMu.RLock()
	res.World = gameState.World
	res.AliveCells = gameState.AliveCells
	res.Turn = gameState.Turn
	stateMu.RUnlock()
	keyCh.Kill <- true
	return
}

func (b *Broker) ClientQuit(req stubs.KeyPressRequest, res *stubs.GameResponse) (err error) {
	stateMu.RLock()
	res.Turn = gameState.Turn
	stateMu.RUnlock()
	keyCh.Quit <- true
	return
}

func (b *Broker) PauseGame(req stubs.KeyPressRequest, res *stubs.GameResponse) (err error) {
	stateMu.Lock()
	gameState.Paused = !gameState.Paused
	res.Turn = gameState.Turn
	res.Paused = gameState.Paused
	if gameState.Paused {
		res.Message = fmt.Sprintf("Current Turn: %v\n", res.Turn)
	} else {
		res.Message = "Continuing\n"
	}
	stateMu.Unlock()
	fmt.Println("Controller Press Pause")
	return
}

func (b *Broker) ReportAliveCells(req stubs.TickerRequest, res *stubs.TickerResponse) (err error) {
	stateMu.RLock()
	res.AliveCells = gameState.AliveCells
	res.Turn = gameState.Turn
	stateMu.RUnlock()
	return
}

func (b *Broker) CloseBroker(req stubs.CloseRequest, res *stubs.CloseResponse) (err error) {
	fmt.Println("Quit Broker")
	quitBroker <- true
	return
}

func (b *Broker) RunGol(req stubs.GameRequest, res *stubs.GameResponse) (err error) {
	fmt.Println("Running Gol")
	p := req.P
	turn := req.Turn

	n := len(workers)

	// Extension: Fault Tolerance
	stateMu.Lock()
	if gameState.Resume {
		turn = gameState.Turn
		req = stubs.GameRequest{
			World: gameState.World,
			P:     req.P,
			Turn:  turn,
		}
		gameState.Resume = false
		fmt.Println("There is an existing Game State, start at turn: ", turn)
	}
	stateMu.Unlock()

	initialise(req, res, turn)

	stateMu.Lock()
	gameState.Save(res)
	stateMu.Unlock()

	for turn < p.Turns {
		select {
		case <-keyCh.Quit:
			stateMu.Lock()
			gameState.Save(res)
			gameState.Resume = true
			gameState.Load(res)
			stateMu.Unlock()
			return
		case <-keyCh.Kill:
			CloseServer()
			res.Kill = true
			return
		default:
			stateMu.Lock()
			if !gameState.Paused {
				turn++
				if n != 1 {
					haloExchange()
				}
				calculateStateAndCells(res, turn)
				gameState.Save(res)
			}
			stateMu.Unlock()
		}
	}
	return
}

func main() {
	pAddr := flag.String("port", "8030", "Port to listen on")
	flag.Parse()

	broker := new(Broker)
	err := rpc.Register(broker)
	util.Check(err)

	listener, _ := net.Listen("tcp", ":"+*pAddr)
	go rpc.Accept(listener)

	<-quitBroker

	defer func(listener net.Listener) {
		err = listener.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(listener)
	os.Exit(0)

}
