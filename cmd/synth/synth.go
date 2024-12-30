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

	reader, enc, _, err := synth.ExampleTones(0)
	if err != nil {
		log.Fatal(err)
	}

	if err := synth.Save(reader, enc, fpath); err != nil {
		log.Fatal(err)
	}
}
