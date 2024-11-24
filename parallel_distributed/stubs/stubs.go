package stubs

import "uk.ac.bris.cs/gameoflife/util"

// Worker methods
var Initialise = "Worker.Initialise"
var HaloExchange = "Worker.HaloExchange"
var CloseServer = "Worker.CloseServer"

// Broker methods
var RunGol = "Broker.RunGol"
var Register = "Broker.Register"
var Ticker = "Broker.ReportAliveCells"
var SaveWorld = "Broker.SaveWorld"
var ClientQuit = "Broker.ClientQuit"
var PauseGame = "Broker.PauseGame"
var ShutDownService = "Broker.ShutDownService"
var CloseBroker = "Broker.CloseBroker"

// Params provides the details of how to run the Game of Life and which image to load.
type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}

type GameRequest struct {
	World [][]byte
	P     Params
	Turn  int
}

type GameResponse struct {
	World      [][]byte
	Turn       int
	AliveCells []util.Cell
	Workers    int
	Message    string
	Kill       bool
	Paused     bool
}

type TickerRequest struct{}

type TickerResponse struct {
	AliveCells []util.Cell
	Turn       int
}

type BrokerRequest struct {
	PartialWorld [][]byte
	P            Params
	StartY       int
	PrevAddr     string
	NextAddr     string
	Workers      int
	IP           string
}

type BrokerResponse struct {
	PartialWorld      [][]byte
	PartialAliveCells []util.Cell
}

type ProcessRequest struct{}

type ProcessResponse struct {
	PartialWorld      [][]byte
	PartialAliveCells []util.Cell
}

type WorkerRequest struct {
	IP string
}

type WorkerResponse struct {
	Message string
}

type KeyPressRequest struct{}
type KeyPressResponse struct {
	World      [][]byte
	AliveCells []util.Cell
	Turn       int
	Workers    int
	Message    string
}

type HaloRequest struct {
}

type HaloResponse struct {
	Top    []byte
	Bottom []byte
}

type CloseRequest struct{}
type CloseResponse struct{}
