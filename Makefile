TARGETS := \
	github.com/modcloth-labs/hookworm \
	github.com/modcloth-labs/hookworm/hookworm-server
VERSION_VAR := github.com/modcloth-labs/hookworm.VersionString
REPO_VERSION := $(shell git describe --always --dirty --tags)
GOBUILD_VERSION_ARGS := -ldflags "-X $(VERSION_VAR) $(REPO_VERSION)"

ADDR := :9988


all: clean test golden README.md

test: build
	go test $(GOBUILD_VERSION_ARGS) -x -v $(TARGETS)

build: deps
	go install $(GOBUILD_VERSION_ARGS) -x $(TARGETS)

deps:
	go get $(GOBUILD_VERSION_ARGS) -x $(TARGETS)
	ruby -rmail/version -e 'Mail::VERSION' || gem install mail --no-ri --no-rdoc

clean:
	rm -rf ./log
	find $${GOPATH%%:*}/pkg -regex '.*modcloth-labs/hookworm.*\.a' -exec rm -v {} \;
	go clean -x $(TARGETS) || true

golden:
	./runtests -v 2>&1 | tee runtests.log

README.md: README.md.in $(wildcard *.go)
	./build-readme $< > $@

serve:
	$${GOPATH%%:*}/bin/hookworm-server -a $(ADDR) -S

todo:
	@grep -R TODO . | grep -v '^./Makefile'


.PHONY: all build clean deps serve test todo
