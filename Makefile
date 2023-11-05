# Makefile for Go Chat Server
.PHONY: all build clean test run

# Binary name for output
BINARY_NAME=chatserver

# Default command to start off the entire build
all: build

# Command to build the server
build:
	@echo "Building..."
	go build -o $(BINARY_NAME) main.go

# Command to clean up the output
clean:
	@echo "Cleaning..."
	go clean
	rm -f $(BINARY_NAME)

# Command to run the server
run: build
	@echo "Running..."
	./$(BINARY_NAME)

# Command to run tests
test:
	@echo "Testing..."
	go test ./...

