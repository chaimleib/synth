TARGET=synth

default: build

build:
	go build -o $(TARGET) ./main.go

test: build
	./$(TARGET)

.PHONY: build test
