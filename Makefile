# Kibble Build System
APP_NAME    := kibble
VERSION     := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME  := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS     := -s -w -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)

.PHONY: all build build-arm64 build-arm build-all run clean test lint size

all: build

build:
	go build -ldflags "$(LDFLAGS)" -o bin/$(APP_NAME) ./cmd/kibble

build-arm64:
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/$(APP_NAME)-linux-arm64 ./cmd/kibble

build-arm:
	GOOS=linux GOARCH=arm GOARM=7 go build -ldflags "$(LDFLAGS)" -o bin/$(APP_NAME)-linux-arm ./cmd/kibble

build-all: build build-arm64 build-arm
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/$(APP_NAME)-linux-amd64 ./cmd/kibble
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/$(APP_NAME)-darwin-arm64 ./cmd/kibble
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/$(APP_NAME)-windows-amd64.exe ./cmd/kibble

run: build
	./bin/$(APP_NAME) -config config.yaml

test:
	go test -v -race -count=1 ./...

clean:
	rm -rf bin/

lint:
	go vet ./...

size: build-arm64
	ls -lh bin/$(APP_NAME)-linux-arm64
