BINARY := keeper
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"
GOFILES := $(shell find . -name '*.go' -not -path './vendor/*')

.PHONY: build install clean test lint run

build:
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/keeper

install: build
	cp bin/$(BINARY) $(GOPATH)/bin/$(BINARY) 2>/dev/null || cp bin/$(BINARY) ~/go/bin/$(BINARY)

clean:
	rm -rf bin/

test:
	go test ./...

lint:
	golangci-lint run ./...

run:
	go run ./cmd/keeper

fmt:
	gofmt -s -w $(GOFILES)

vet:
	go vet ./...

# Development helpers
dev-init:
	go run ./cmd/keeper init

dev-add:
	go run ./cmd/keeper add

dev-list:
	go run ./cmd/keeper list

dev-dashboard:
	go run ./cmd/keeper dashboard

dev-doctor:
	go run ./cmd/keeper doctor
