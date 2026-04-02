.PHONY: build test lint clean install

BINARY := openapi-splitter
MODULE  := github.com/akyrey/openapi-splitter

build:
	go build -o $(BINARY) .

test:
	go test ./... -count=1

lint:
	golangci-lint run ./...

clean:
	rm -f $(BINARY)

install:
	go install .
