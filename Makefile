TARGETS := \
	github.com/modcloth-labs/hookworm \
	github.com/modcloth-labs/hookworm/hookworm-server
VERSION_VAR := github.com/modcloth-labs/hookworm.VersionString
REPO_VERSION := $(shell git describe --always --dirty --tags)
GOBUILD_VERSION_ARGS := -ldflags "-X $(VERSION_VAR) $(REPO_VERSION)"

ADDR := :9988


all: test README.md

test: build
	go test $(GOBUILD_VERSION_ARGS) -x -v $(TARGETS)

build: deps
	go install $(GOBUILD_VERSION_ARGS) -x $(TARGETS)

deps:
	go get $(GOBUILD_VERSION_ARGS) -x $(TARGETS)

README.md: README.md.in $(wildcard *.go)
	./build-readme $< > $@

serve:
	$${GOPATH%%:*}/bin/hookworm-server -a $(ADDR) -S


.PHONY: all test build deps
