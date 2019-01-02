PKG = github.com/k1LoW/evry
COMMIT = $$(git describe --tags --always)
OSNAME=${shell uname -s}
ifeq ($(OSNAME),Darwin)
	DATE = $$(gdate --utc '+%Y-%m-%d_%H:%M:%S')
else
	DATE = $$(date --utc '+%Y-%m-%d_%H:%M:%S')
endif

BUILD_LDFLAGS = -X $(PKG).commit=$(COMMIT) -X $(PKG).date=$(DATE)
RELEASE_BUILD_LDFLAGS = -s -w $(BUILD_LDFLAGS)

GO ?= GO111MODULE=on go

default: test

ci: build test

test: build
	$(GO) test ./... -coverprofile=coverage.txt -covermode=count

build:
	$(GO) build -ldflags="$(BUILD_LDFLAGS)"
