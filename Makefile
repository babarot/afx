BINARY_NAME := afx
VERSION := $(shell cat VERSION)
GIT_SHA := $(shell git rev-parse --verify --short HEAD)
GIT_TAG := $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null || echo "canary")
LDFLAGS := "-X github.com/babarot/afx/cmd.Version=v$(VERSION) -X github.com/babarot/afx/cmd.BuildTag=$(GIT_TAG) -X github.com/babarot/afx/cmd.BuildSHA=$(GIT_SHA)"

all: build

test:
	go test -race -coverprofile=coverage.out ./...

lint:
	golangci-lint run ./...

build:
	go build -ldflags $(LDFLAGS) -trimpath -o $(BINARY_NAME) .

install:
	go install -ldflags $(LDFLAGS) .

clean:
	rm -f $(BINARY_NAME) coverage.out

docs:
	@python3 -m pip install --upgrade pip
	@python3 -m pip install -r ./docs/requirements.txt
	@python3 -m mkdocs serve

.PHONY: all test lint build install clean docs
