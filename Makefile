TARGET=synth

all: synth play

play:
	go run ./cmd/play

synth:
	go run ./cmd/synth beep.wav

.PHONY: build test
