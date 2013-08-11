TARGETS := hookworm
VERSION_VAR := hookworm.VersionString
REPO_VERSION := $(shell git describe --always --dirty --tags)

REV_VAR := hookworm.RevisionString
REPO_REV := $(shell git rev-parse --sq HEAD)

GOBUILD_VERSION_ARGS := -ldflags "-X $(VERSION_VAR) $(REPO_VERSION) -X $(REV_VAR) $(REPO_REV)"

ADDR := :9988

all: clean test golden README.md

test: build
	go test $(GOBUILD_VERSION_ARGS) -x -v $(TARGETS)

build: deps
	go install $(GOBUILD_VERSION_ARGS) -x $(TARGETS)
	go build -o $${GOPATH%%:*}/bin/hookworm-server $(GOBUILD_VERSION_ARGS) ./hookworm-server

deps:
	if [ ! -L $${GOPATH%%:*}/src/hookworm ] ; then gvm linkthis ; fi
	gem query --local | grep -Eq '^mail\b.*\b2\.5\.4\b'  || \
	  gem install mail -v 2.5.4 --no-ri --no-rdoc
	gem query --local | grep -Eq '^fakesmtpd\b.*\b0\.1\.0\b'  || \
	  gem install fakesmtpd -v 0.1.0 --no-ri --no-rdoc

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
	@grep -n -R TODO . | grep -v -E '^(./Makefile|./.git)'

.PHONY: all build clean deps serve test todo golden
