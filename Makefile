.PHONY: pkg clean release release

VERSION=5.0.0
RELEASE=rc1
GENERATOR=RPM
APP_NAME?=agentd
CGO_ENABLED?=1
OBJ=$(APP_NAME)
RPM_FILE=$(APP_NAME)-$(VERSION)-$(RELEASE).$(shell uname -s).$(shell uname -m).rpm
DEP_OBJS=$(shell find . -name "*.go" -type f -not -path "./vendor/*" -a -not -path "./staging/*")
TARGETS=directories build/agentd
GO_BUILD_LDFLAGS_CMD=$(abspath ./scripts/go-build-ldflags.sh)
GO_BUILD_LDFLAGS=$(shell $(GO_BUILD_LDFLAGS_CMD) LDFLAG)

all: $(TARGETS)

pkg: build/$(RPM_FILE)

pkgs: $(TARGETS)
	docker run --rm \
		-v $(PWD):/src \
		--name golang-cross-builder \
		--hostname golang-cross-builder \
		-e GO_BUILD_LDFLAGS='$(GO_BUILD_LDFLAGS)' \
		-e VERSION=$(VERSION) \
		-e RELEASE=$(RELEASE) \
		-it ghcr.io/gythialy/golang-cross-builder:v1.16.2 \
		/src/scripts/pkgs.sh

	#VERSION=$(VERSION) RELEASE=$(RELEASE) \
	#GO_BUILD_LDFLAGS='$(GO_BUILD_LDFLAGS)' \
	#./scripts/package.sh


build/_$(APP_NAME)/Makefile: $(TARGETS) Makefile
	mkdir -p build/_$(APP_NAME); cd build/_$(APP_NAME); cmake ../.. -DAPP_NAME=$(APP_NAME) -DCPACK_PACKAGE_VERSION="$(VERSION)" -DCPACK_PACKAGE_RELEASE="$(RELEASE)" -DCPACK_GENERATOR="$(GENERATOR)"

build/agentd: $(DEP_OBJS)
	GO111MODULE=on \
	go build -ldflags '$(GO_BUILD_LDFLAGS)' \
	-o $@ ./cmd/agentd && \
	$@ version

directories: build

build:
	mkdir -p build

build/$(RPM_FILE): build/_$(APP_NAME)/Makefile $(TARGETS)
	cd build/_$(APP_NAME) && make package && cp -af $(RPM_FILE) ../
	rpm -pql build/$(RPM_FILE)

clean:
	rm -rf build


release:
	VERSION=$(VERSION) RELEASE=$(RELEASE) ./scripts/release.sh

devrun: build/agentd
	@echo "./build/agentd start --config ./etc/agentd.yml -v 10 2>&1"

dev: build/agentd
	APP_NAME=agentd watcher -logtostderr \
		 -v 10 -e build -e .git -e docs \
		 -e plugins -e tmp -e vendor -e staging -f .go -d 1000
