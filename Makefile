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

all: clean test README.md

test: build fmtpolice testdeps coverage.html

coverage.html: coverage.out
	$(GO) tool cover -html=$^ -o $@

coverage.out:
	$(GO) test -covermode=count -coverprofile=$@ $(GOBUILD_LDFLAGS) $(GO_TAG_ARGS) -x -v $(HOOKWORM_PACKAGE)
	$(GO) tool cover -func=$@

testdeps:
	$(GO) test -i $(GOBUILD_LDFLAGS) $(GO_TAG_ARGS) -x -v $(TARGETS)

build: deps
	$(GO) install $(GOBUILD_LDFLAGS) $(GO_TAG_ARGS) -x $(TARGETS)

deps: public
	if [ ! -e $${GOPATH%%:*}/src/$(HOOKWORM_PACKAGE) ] ; then \
		mkdir -p $${GOPATH%%:*}/src/github.com/modcloth-labs ; \
		ln -sv $(PWD) $${GOPATH%%:*}/src/$(HOOKWORM_PACKAGE) ; \
	fi
	bundle install
	$(GO) get -x $(GOBUILD_LDFLAGS) $(GO_TAG_ARGS) -x $(TARGETS)
	$(GODEP) restore

clean:
	rm -rf ./log ./coverage.out
	$(GO) clean -x $(TARGETS) || true
	if [ -d $${GOPATH%%:*}/pkg ] ; then \
		find $${GOPATH%%:*}/pkg -name '*hookworm*' -exec rm -v {} \; ; \
	fi

save:
	$(GODEP) save -copy=false $(HOOKWORM_PACKAGE)

container:
	$(DOCKER) build -t quay.io/modcloth/hookworm:$(REPO_VERSION) $(BUILD_FLAGS) .

fmtpolice:
	set -e; for f in $(shell git ls-files '*.go'); do gofmt $$f | diff -u $$f - ; done

public:
	mkdir -p $@

README.md: README.in.md $(shell git ls-files '*.go') $(shell git ls-files 'worm.d/*.*')
	./build-readme < $< > $@

serve:
	$${GOPATH%%:*}/bin/hookworm-server -a $(ADDR)

todo:
	@grep -n -R TODO . | grep -v -E '^(./Makefile|./.git)'

.PHONY: all build clean container deps serve test fmtpolice todo golden
