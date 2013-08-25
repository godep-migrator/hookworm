TARGETS := hookworm

VERSION_VAR := hookworm.VersionString
REPO_VERSION := $(shell git describe --always --dirty --tags)

REV_VAR := hookworm.RevisionString
REPO_REV := $(shell git rev-parse --sq HEAD)

GO_TAG_ARGS ?= -tags full
TAGS_VAR := hookworm.BuildTags
GOBUILD_LDFLAGS := -ldflags "-X $(VERSION_VAR) $(REPO_VERSION) -X $(REV_VAR) $(REPO_REV) -X $(TAGS_VAR) '$(GO_TAG_ARGS)' "

ADDR := :9988

all: clean test golden README.md

test: build
	go test $(GOBUILD_LDFLAGS) $(GO_TAG_ARGS) -x -v $(TARGETS)

build: deps
	go install $(GOBUILD_LDFLAGS) $(GO_TAG_ARGS) -x $(TARGETS)
	go build -o $${GOPATH%%:*}/bin/hookworm-server $(GOBUILD_LDFLAGS) $(GO_TAG_ARGS) ./hookworm-server

deps: fakesmtpd mtbb
	if [ ! -L $${GOPATH%%:*}/src/hookworm ] ; then gvm linkthis ; fi
	gem query --local | grep -Eq '^mail\b.*\b2\.5\.4\b'  || \
	  gem install mail -v 2.5.4 --no-ri --no-rdoc

clean:
	rm -rf ./log ./.mtbb-artifacts/ ./tests.log
	go clean -x $(TARGETS) || true
	if [ -d $${GOPATH%%:*}/pkg ] ; then \
		find $${GOPATH%%:*}/pkg -name '*hookworm*' -exec rm -v {} \; ; \
	fi

golden:
	./mtbb -v 2>&1 | tee tests.log

README.md: README.in.md $(wildcard *.go)
	ruby -e "exe = \"#{ENV['GOPATH'].split(':').first}/bin/hookworm-server\" ; \
	  puts \$$<.read.sub(/___USAGE___/, \`#{exe} -h 2>&1\`.chomp)" < $< > $@

fakesmtpd:
	curl -s -o $@ https://raw.github.com/modcloth-labs/fakesmtpd/v0.3.0/lib/fakesmtpd/server.rb
	chmod +x $@

mtbb:
	curl -s -o $@ https://raw.github.com/modcloth-labs/mtbb/v0.1.0/lib/mtbb.rb
	chmod +x $@

serve:
	$${GOPATH%%:*}/bin/hookworm-server -a $(ADDR) -S

todo:
	@grep -n -R TODO . | grep -v -E '^(./Makefile|./.git)'

.PHONY: all build clean deps serve test todo golden
