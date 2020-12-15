.DEFAULT_GOAL := run

build:
	go build -o bin/scheduler main.go
	go build -o bin/executor ./executor

run: build
	./bin/scheduler