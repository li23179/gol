// Example useage:
// go test -run a$ -bench "Gol/512x512x1000-<no of workers>" -timeout 100s
package main

import (
	"fmt"
	"os"
	"testing"
	"uk.ac.bris.cs/gameoflife/gol"
	"uk.ac.bris.cs/gameoflife/stubs"
)

const benchLength = 1000

func BenchmarkGol(b *testing.B) {
	threads := 3    // REMEMBER TO SET THIS MANUALLY
	os.Stdout = nil // Disable all program output apart from benchmark results
	p := stubs.Params{
		Turns:       benchLength,
		Threads:     threads,
		ImageWidth:  512,
		ImageHeight: 512,
	}
	name := fmt.Sprintf("%dx%dx%d-%d", p.ImageWidth, p.ImageHeight, p.Turns, p.Threads)
	b.Run(name, func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			events := make(chan gol.Event)
			go gol.Run(p, events, nil)
			for range events {
			}
		}
	})
}
