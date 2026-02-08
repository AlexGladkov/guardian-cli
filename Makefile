.PHONY: build test lint fmt release docker clean install

BINARY_NAME=guardian
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/guardian

test:
	go test -v -race -coverprofile=coverage.out ./...

lint:
	golangci-lint run ./...

fmt:
	gofmt -s -w .

release:
	goreleaser release --clean

docker:
	docker build -t guardian:$(VERSION) .

clean:
	rm -rf bin/ dist/ coverage.out

install:
	go install $(LDFLAGS) ./cmd/guardian
