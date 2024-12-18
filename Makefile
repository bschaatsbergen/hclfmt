VERSION=$(shell git tag | tail -1)

default: fmt lint build

all: fmt lint build

fmt:
	@echo "==> Fixing source code with gofmt..."
	find . -name '*.go' | grep -v vendor | xargs gofmt -s -w

lint: fmt
	@echo "==> Checking source code against linters..."
	golangci-lint run

build: fmt
	@echo "==> building..."
	go build -o $(shell go env GOPATH)/bin/hclfmt .

test:
	@echo "==> Thank you for running the tests!"
	go test -race -parallel 8 ./...

.PHONY: fmt lint build
