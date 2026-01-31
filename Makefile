.PHONY: build install test test-unit test-integration test-e2e test-coverage clean run fmt lint

fmt:
	gofmt -w .

lint: fmt
	go vet ./...

build: fmt
	go build -o bin/grund .

install:
	go install .

test:
	go test ./...

test-unit:
	go test ./internal/domain/... -v

test-integration:
	go test ./internal/application/... -v

test-e2e:
	go test ./test/integration/... -v

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-race:
	go test -race ./...

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

run:
	go run .
