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

test: build fmtpolice
	go test -race $(GOBUILD_LDFLAGS) $(GO_TAG_ARGS) -x -v $(TARGETS)

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

distclean: clean
	rm -f mtbb fakesmtpd

golden:
	./mtbb -v 2>&1 | tee tests.log

fmtpolice:
	set -e; for f in $(shell git ls-files '*.go'); do gofmt $$f | diff -u $$f - ; done

README.md: README.in.md $(shell git ls-files '*.go') $(shell git ls-files 'worm.d/*.*')
	./build-readme < $< > $@

fakesmtpd:
	curl -s -o $@ https://raw.github.com/modcloth-labs/fakesmtpd/v0.3.1/lib/fakesmtpd/server.rb
	chmod +x $@

mtbb:
	curl -s -o $@ https://raw.github.com/modcloth-labs/mtbb/v0.1.1/lib/mtbb.rb
	chmod +x $@

serve:
	$${GOPATH%%:*}/bin/hookworm-server -a $(ADDR) -S

todo:
	@grep -n -R TODO . | grep -v -E '^(./Makefile|./.git)'

.PHONY: all build clean distclean deps serve test fmtpolice todo golden
