.PHONY: build install test test-coverage coverage coverage-view

# Build the binary
build:
	go build -o bin/kiwi cmd/kiwi/main.go

# Install locally
install:
	go install ./cmd/kiwi

# Run all tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -cover ./...

# Generate coverage report
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# View coverage report in browser
coverage-view:
	go tool cover -html=coverage.out 