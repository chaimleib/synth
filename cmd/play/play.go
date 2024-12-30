package main

import (
	"log"
	"time"

	"github.com/chaimleib/synth"
)

func main() {
	chunkDuration := 100 * time.Millisecond
	reader, enc, chunkSize, err := synth.ExampleTones(chunkDuration)
	if err != nil {
		log.Fatal(err)
	}
	if err := synth.Play(reader, enc, chunkSize); err != nil {
		log.Fatal(err)
	}
}
