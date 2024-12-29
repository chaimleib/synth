package main

import (
	"log"

	"github.com/chaimleib/synth"
)

func main() {
	reader, err := synth.ExampleTones()
	if err != nil {
		log.Fatal(err)
	}
	if err := synth.Play(reader); err != nil {
		log.Fatal(err)
	}
}
