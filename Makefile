HOOKWORM_PACKAGE := github.com/modcloth-labs/hookworm
TARGETS := \
  $(HOOKWORM_PACKAGE) \
  $(HOOKWORM_PACKAGE)/hookworm-server

VERSION_VAR := $(HOOKWORM_PACKAGE).VersionString
REPO_VERSION := $(shell git describe --always --dirty --tags)

REV_VAR := $(HOOKWORM_PACKAGE).RevisionString
REPO_REV := $(shell git rev-parse --sq HEAD)

GO ?= go
GODEP ?= godep
GO_TAG_ARGS ?= -tags full
TAGS_VAR := $(HOOKWORM_PACKAGE).BuildTags
GOBUILD_LDFLAGS := -ldflags "-X $(VERSION_VAR) $(REPO_VERSION) -X $(REV_VAR) $(REPO_REV) -X $(TAGS_VAR) '$(GO_TAG_ARGS)' "

DOCKER ?= sudo docker
BUILD_FLAGS ?= -no-cache=true -rm=true

ADDR := :9988

all: clean test golden README.md

test: build fmtpolice rubocop
	$(GO) test -i $(GOBUILD_LDFLAGS) $(GO_TAG_ARGS) -x -v $(TARGETS)
	$(GO) test -race $(GOBUILD_LDFLAGS) $(GO_TAG_ARGS) -x -v $(TARGETS)

build: deps
	$(GO) install $(GOBUILD_LDFLAGS) $(GO_TAG_ARGS) -x $(TARGETS)

deps: fakesmtpd mtbb public
	if [ ! -e $${GOPATH%%:*}/src/$(HOOKWORM_PACKAGE) ] ; then \
		mkdir -p $${GOPATH%%:*}/src/github.com/modcloth-labs ; \
		ln -sv $(PWD) $${GOPATH%%:*}/src/$(HOOKWORM_PACKAGE) ; \
	fi
	gem query --local | grep -Eq '^mail\b.*\b2\.5\.4\b'  || \
		gem install mail -v '2.5.4' --no-ri --no-rdoc
	gem query --local | grep -Eq '^rubocop\b.*\b0\.17\.0\b'  || \
		gem install rubocop -v '~> 0.17.0' --no-ri --no-rdoc
	gem query --local | grep -Eq '^hookworm-base\b.*\b0\.1\.0\b'  || \
		gem install hookworm-base -v '~> 0.1.0' --no-ri --no-rdoc
	$(GO) get -x $(GOBUILD_LDFLAGS) $(GO_TAG_ARGS) -x $(TARGETS)
	$(GODEP) restore

clean:
	rm -rf ./log ./.mtbb-artifacts/ ./tests.log
	$(GO) clean -x $(TARGETS) || true
	if [ -d $${GOPATH%%:*}/pkg ] ; then \
		find $${GOPATH%%:*}/pkg -name '*hookworm*' -exec rm -v {} \; ; \
	fi

save:
	$(GODEP) save -copy=false $(HOOKWORM_PACKAGE)

container:
	$(DOCKER) build -t quay.io/modcloth/hookworm:$(REPO_VERSION) $(BUILD_FLAGS) .

distclean: clean
	rm -f mtbb fakesmtpd

golden:
	./mtbb -v 2>&1 | tee tests.log

fmtpolice:
	set -e; for f in $(shell git ls-files '*.go'); do gofmt $$f | diff -u $$f - ; done

rubocop:
	rubocop --config .rubocop.yml --format simple

public:
	mkdir -p $@

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

.PHONY: all build clean container distclean deps serve test fmtpolice todo golden
