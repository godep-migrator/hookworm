TARGETS := hookworm
VERSION_VAR := hookworm.VersionString
REPO_VERSION := $(shell git describe --always --dirty --tags)
GOBUILD_VERSION_ARGS := -ldflags "-X $(VERSION_VAR) $(REPO_VERSION)"

ADDR := :9988

all: clean test golden README.md

test: build
	go test $(GOBUILD_VERSION_ARGS) -x -v $(TARGETS)

build: deps
	go install $(GOBUILD_VERSION_ARGS) -x $(TARGETS)
	go build -o $${GOPATH%%:*}/bin/hookworm-server ./hookworm-server

deps: gvm_check
	if [ ! -L $${GOPATH%%:*}/src/hookworm ] ; then gvm linkthis ; fi
	ruby -rmail/version -e 'Mail::VERSION' 2>/dev/null || gem install mail --no-ri --no-rdoc

clean:
	rm -rf ./log
	go clean -x $(TARGETS) || true
	if [ -d $${GOPATH%%:*}/pkg ] ; then \
		find $${GOPATH%%:*}/pkg -name '*hookworm*' -exec rm -v {} \; ; \
	fi

golden:
	./runtests -v 2>&1 | tee runtests.log

README.md: README.md.in $(wildcard *.go)
	./build-readme $< > $@

serve:
	$${GOPATH%%:*}/bin/hookworm-server -a $(ADDR) -S

todo:
	@grep -R TODO . | grep -v '^./Makefile'

gvm_check:
	which gvm

.PHONY: all build clean deps serve test todo golden
