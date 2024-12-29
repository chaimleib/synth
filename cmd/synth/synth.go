package main

import (
	"log"
	"os"

	"github.com/chaimleib/synth"
)

func main() {
	if len(os.Args[1:]) != 1 {
		log.Fatal("expected a filepath argument")
	}
	fpath := os.Args[1]

	reader, err := synth.ExampleTones()
	if err != nil {
		log.Fatal(err)
	}

	if err := synth.Save(reader, fpath); err != nil {
		log.Fatal(err)
	}
}
