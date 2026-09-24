.PHONY: all build run test clean

BINARY_NAME=bin/cloud

all: build

build: ## Builds the CLI binary at bin/cloud
	@mkdir -p bin
	go build -o $(BINARY_NAME) main.go

run: build ## Builds and runs the CLI
	./$(BINARY_NAME)

test: ## Runs all unit tests
	go test -v ./...

clean: ## Cleans generated binaries
	rm -rf bin/
