.PHONY: all build run clean test setup

all: build

build:
	go build -o elo-cli ./cmd/elo-cli

run:
	go run ./cmd/elo-cli

test:
	go test -v ./...

clean:
	rm -f elo-cli
	rm -f data/*.db
	rm -f data/matches-processed/*

setup:
	mkdir -p data/matches-pending
	mkdir -p data/matches-processed
	mkdir -p data/matches-failed
	mkdir -p docs

deps:
	go mod download
	go mod tidy

install: build
	cp elo-cli $(GOPATH)/bin/
