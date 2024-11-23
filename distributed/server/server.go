package main

import (
	"flag"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

func checkError(err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
}

const dead byte = 0
const live byte = 255

var QuitServer = make(chan bool)

func CalculateLiveNeighbour(p stubs.Params, x, y, maxY int, immutableWorld func(int, int) byte) int {
	counter := 0
	relativeDir := [8][2]int{
		{0, 1}, {0, -1},
		{1, 0}, {-1, 0},
		{-1, -1}, {-1, 1},
		{1, 1}, {1, -1},
	}

	for _, dir := range relativeDir {
		ny := (dir[1] + maxY + y) % maxY
		nx := (dir[0] + p.ImageWidth + x) % p.ImageWidth

		if immutableWorld(ny, nx) == live {
			counter++
		}
	}
	return counter
}

func CalculateNextState(p stubs.Params, startY, endY, maxY int, immutableWorld func(int, int) byte) [][]byte {
	// TODO : Implement a parallel version for workers
	newWorld := util.MakeWorld(p.ImageWidth, endY-startY)
	for y := startY; y < endY; y++ {
		j := y - startY // adjust the column
		for x := 0; x < p.ImageWidth; x++ {
			counter := CalculateLiveNeighbour(p, x, y, maxY, immutableWorld)
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
func CalculateAliveCells(p stubs.Params, startY, endY, offSetY int, immutableWorld func(int, int) byte) []util.Cell {
	var cells []util.Cell
	for y := startY; y < endY; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			if immutableWorld(y, x) == live {
				cells = append(cells, util.Cell{X: x, Y: y + offSetY})
			}
		}
	}
	return cells
}

// StateWorker worldWorker work on the same board
// use goroutine to distribute the worker on different section and export via channel
func StateWorker(p stubs.Params, startY, endY, maxY int, immutableWorld func(int, int) byte, worldCh chan<- [][]byte) {
	partialWorld := CalculateNextState(p, startY, endY, maxY, immutableWorld)
	worldCh <- partialWorld
	//fmt.Println("finish worldCh")
}

func AliveCellWorker(p stubs.Params, startY, endY, offSetY int, immutableWorld func(int, int) byte, aliveCellCh chan<- []util.Cell) {
	partialAliveCells := CalculateAliveCells(p, startY, endY, offSetY, immutableWorld)
	aliveCellCh <- partialAliveCells
}

// DelegateStateWork accumulate workload for each worker for each turn
func DelegateStateWork(p stubs.Params, maxY int, immutableWorld func(int, int) byte, worldChs []chan [][]byte, finishWorldCh chan<- [][]byte) {
	baseWorkload := maxY / p.Threads
	extraWorkerThreads := maxY % p.Threads

	startY := 0
	var finishWorld [][]byte

	for t := 0; t < p.Threads; t++ {
		workload := baseWorkload
		if t < extraWorkerThreads {
			workload++
		}
		endY := startY + workload
		go StateWorker(p, startY, endY, maxY, immutableWorld, worldChs[t])
		startY = endY
	}

	for t := 0; t < p.Threads; t++ {
		finishWorld = append(finishWorld, <-worldChs[t]...)
	}
	finishWorldCh <- finishWorld
}

func DelegateCellWork(p stubs.Params, maxY, offSetY int, immutableWorld func(int, int) byte, aliveCellChs []chan []util.Cell, finishCellsCh chan<- []util.Cell) {
	baseWorkload := maxY / p.Threads
	extraWorkerThreads := maxY % p.Threads

	startY := 0

	var finishCells []util.Cell

	for t := 0; t < p.Threads; t++ {
		workload := baseWorkload
		if t < extraWorkerThreads {
			workload++
		}
		endY := startY + workload
		go AliveCellWorker(p, startY, endY, offSetY, immutableWorld, aliveCellChs[t])
		startY = endY
	}

	for t := 0; t < p.Threads; t++ {
		finishCells = append(finishCells, <-aliveCellChs[t]...)
	}
	finishCellsCh <- finishCells
}

type HaloRegion struct {
	TopRow    []byte
	BottomRow []byte
}

type Worker struct {
	P          stubs.Params
	World      [][]byte
	AliveCells []util.Cell
	StartY     int
	// Use for sending Halo Regions
	CurrentTop    []byte
	CurrentBottom []byte
	HaloRegion    HaloRegion

	PrevAddr string
	NextAddr string
	Workers  int
	IP       string

	StateChannels          []chan [][]byte
	AliveCellChannels      []chan []util.Cell
	ResultStateChannel     chan [][]byte
	ResultAliveCellChannel chan []util.Cell
}

func (w *Worker) receiveHaloTop(prev string, req stubs.HaloRequest, res *stubs.HaloResponse) {
	if prev != w.IP {
		client, _ := rpc.Dial("tcp", prev)
		_ = client.Call("Worker.SendBottomRegion", req, res)
		w.HaloRegion.TopRow = res.Top
	} else {
		w.HaloRegion.TopRow = make([]byte, len(w.CurrentBottom))
		copy(w.HaloRegion.TopRow, w.CurrentTop)
	}
}

func (w *Worker) SendBottomRegion(req stubs.HaloRequest, res *stubs.HaloResponse) (err error) {
	res.Top = w.CurrentBottom
	return err
}

func (w *Worker) receiveHaloBottom(next string, req stubs.HaloRequest, res *stubs.HaloResponse) {
	if next != w.IP {
		client, _ := rpc.Dial("tcp", next)
		_ = client.Call("Worker.SendTopRegion", req, res)
		w.HaloRegion.BottomRow = res.Bottom
	} else {
		w.HaloRegion.BottomRow = make([]byte, len(w.CurrentTop))
		copy(w.HaloRegion.BottomRow, w.CurrentTop)
	}
}

func (w *Worker) SendTopRegion(req stubs.HaloRequest, res *stubs.HaloResponse) (err error) {
	res.Bottom = w.CurrentTop
	return err
}

func (w *Worker) HaloExchange(req stubs.BrokerRequest, res *stubs.BrokerResponse) (err error) {
	haloRequest := stubs.HaloRequest{}

	haloResponse := new(stubs.HaloResponse)
	fmt.Println("Exchanging")

	w.receiveHaloTop(w.PrevAddr, haloRequest, haloResponse)
	w.receiveHaloBottom(w.NextAddr, haloRequest, haloResponse)
	return
}

func (w *Worker) addHaloRegion() {
	var world [][]byte
	world = append(world, w.HaloRegion.TopRow)
	world = append(world, w.World...)
	world = append(world, w.HaloRegion.BottomRow)
	w.World = world
}

func (w *Worker) filterHaloRegion() {
	p := w.P
	world := util.MakeWorld(p.ImageWidth, len(w.World)-2)
	for i := range world {
		world[i] = w.World[i+1]
	}
	w.World = world
}

func (w *Worker) haloRegionReset() {
	w.HaloRegion = HaloRegion{TopRow: nil, BottomRow: nil}
}

func (w *Worker) Initialise(req stubs.BrokerRequest, res *stubs.BrokerResponse) (err error) {
	fmt.Println("initialise")
	w.P = req.P
	w.IP = req.IP
	w.World = req.PartialWorld
	w.StartY = req.StartY

	w.AliveCellChannels = util.MakeCellWorkerChannels(w.P.Threads)
	w.StateChannels = util.MakeStateWorkerChannels(w.P.Threads)
	w.ResultStateChannel = util.MakeNextStateChannel()
	w.ResultAliveCellChannel = util.MakeNextAliveCellChannel()

	maxY := len(w.World)

	immutableWorld := util.MakeImmutableWorld(w.World)

	go DelegateCellWork(w.P, maxY, w.StartY, immutableWorld, w.AliveCellChannels, w.ResultAliveCellChannel)
	w.AliveCells = <-w.ResultAliveCellChannel

	w.NextAddr = req.NextAddr
	w.PrevAddr = req.PrevAddr

	w.CurrentTop = w.World[0]
	w.CurrentBottom = w.World[len(w.World)-1]
	w.Workers = req.Workers

	res.PartialWorld = w.World
	res.PartialAliveCells = w.AliveCells

	fmt.Println("done")
	return
}

func (w *Worker) ProcessTurn(req stubs.ProcessRequest, res *stubs.ProcessResponse) (err error) {
	fmt.Println("Process turn start")
	if w.Workers == 1 {
		maxY := len(w.World)
		fmt.Println(maxY)
		immutableWorld := util.MakeImmutableWorld(w.World)

		go DelegateStateWork(w.P, maxY, immutableWorld, w.StateChannels, w.ResultStateChannel)
		w.World = <-w.ResultStateChannel

		go DelegateCellWork(w.P, maxY, w.StartY, immutableWorld, w.AliveCellChannels, w.ResultAliveCellChannel)
		immutableWorld = util.MakeImmutableWorld(w.World)
		go DelegateCellWork(w.P, maxY, 0, immutableWorld, w.AliveCellChannels, w.ResultAliveCellChannel)
		w.AliveCells = <-w.ResultAliveCellChannel

	} else {
		w.addHaloRegion()
		maxY := len(w.World)
		immutableWorld := util.MakeImmutableWorld(w.World)

		go DelegateStateWork(w.P, maxY, immutableWorld, w.StateChannels, w.ResultStateChannel)
		w.World = <-w.ResultStateChannel

		w.filterHaloRegion()
		w.haloRegionReset()

		w.CurrentTop = w.World[0]
		w.CurrentBottom = w.World[len(w.World)-1]

		maxY = len(w.World)
		immutableWorld = util.MakeImmutableWorld(w.World)

		go DelegateCellWork(w.P, maxY, w.StartY, immutableWorld, w.AliveCellChannels, w.ResultAliveCellChannel)
		w.AliveCells = <-w.ResultAliveCellChannel
	}

	res.PartialWorld = w.World
	res.PartialAliveCells = w.AliveCells

	fmt.Println("Process turn ends")
	return
}

func (w *Worker) CloseServer(req stubs.CloseRequest, res *stubs.CloseResponse) (err error) {
	QuitServer <- true
	return
}
func main() {
	pAddr := flag.String("port", "127.0.0.1:8050", "Port to listen on")
	bAddr := flag.String("broker", "127.0.0.1:8030", "Access to broker instance")
	flag.Parse()

	golWorker := new(Worker)
	err := rpc.Register(golWorker)

	if err != nil {
		fmt.Println("Can't Register GolEngine for rpc")
		return
	}

	listener, err := net.Listen("tcp", *pAddr)
	client, err := rpc.Dial("tcp", *bAddr)

	req := stubs.WorkerRequest{IP: *pAddr}
	res := new(stubs.WorkerResponse)
	_ = client.Call(stubs.Register, req, res)
	fmt.Println(res.Message)

	fmt.Println("Worker", *pAddr, "is on work")
	checkError(err)

	defer func(listener net.Listener) {
		e := listener.Close()
		if e != nil {
			checkError(e)
		}
	}(listener)

	go rpc.Accept(listener)

	<-QuitServer
	os.Exit(0)
}
