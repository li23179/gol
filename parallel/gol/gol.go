package gol

// Params provides the details of how to run the Game of Life and which image to load.
type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}

// Run starts the processing of Game of Life. It should initialise channels and goroutines.
func Run(p Params, events chan<- Event, keyPresses <-chan rune) {

	//	TODO: Put the missing channels in here.
	fileCh := make(chan string)
	outputCh := make(chan uint8)
	inputCh := make(chan uint8)

	ioCommand := make(chan ioCommand)
	ioIdle := make(chan bool)

	ioChannels := ioChannels{
		command:  ioCommand,
		idle:     ioIdle,
		filename: fileCh,
		output:   outputCh,
		input:    inputCh,
	}
	go startIo(p, ioChannels)

	distributorChannels := distributorChannels{
		events:     events,
		ioCommand:  ioCommand,
		ioIdle:     ioIdle,
		ioFilename: fileCh,
		ioOutput:   outputCh,
		ioInput:    inputCh,
		keyPressCh: keyPresses,
	}
	distributor(p, distributorChannels)
}
