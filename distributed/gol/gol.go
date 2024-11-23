package gol

import "uk.ac.bris.cs/gameoflife/stubs"

// Run starts the processing of Game of Life. It should initialise channels and goroutines.
func Run(p stubs.Params, events chan<- Event, keyPresses <-chan rune) {

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
