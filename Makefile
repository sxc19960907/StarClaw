.PHONY: build test clean install lint fmt coverage

BINARY_NAME=starclaw
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X github.com/starclaw/starclaw/cmd.Version=${VERSION}"

build:
	go build ${LDFLAGS} -o ${BINARY_NAME} .

test:
	go test -v ./...

test-race:
	go test -race -v ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

fmt:
	go fmt ./...

lint:
	golangci-lint run ./...

clean:
	go clean
	rm -f ${BINARY_NAME}
	rm -f coverage.out coverage.html

install:
	go install ${LDFLAGS} .

dev:
	go run ${LDFLAGS} . $(ARGS)

# Dependencies
deps:
	go mod tidy
	go mod download

update:
	go get -u ./...
	go mod tidy

# Build for all platforms
build-all:
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-linux-arm64 .
