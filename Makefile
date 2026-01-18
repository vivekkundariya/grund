.PHONY: build install test clean

build:
	go build -o bin/grund ./cmd/grund

install:
	go install ./cmd/grund

test:
	go test ./...

clean:
	rm -rf bin/

run:
	go run ./cmd/grund
