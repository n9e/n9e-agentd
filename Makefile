.PHONY: pkg clean release release dev devrun

VERSION=5.0.0
RELEASE=rc8
GENERATOR=RPM
APP_NAME?=n9e-agentd
CGO_ENABLED?=1
OBJ=$(APP_NAME)
RPM_FILE=$(APP_NAME)-$(VERSION)-$(RELEASE).$(shell uname -s).$(shell uname -m).rpm
DEP_OBJS=$(shell find . -name "*.go" -type f -not -path "./vendor/*" -a -not -path "./staging/*") \
	 pkg/data/resources.go
TARGETS?=directories build/n9e-agentd # build/agentdctl
GO_BUILD_LDFLAGS_CMD=$(abspath ./scripts/go-build-ldflags.sh)
GO_BUILD_LDFLAGS=$(shell $(GO_BUILD_LDFLAGS_CMD) LDFLAG)

all: $(TARGETS)

include ./scripts/objs.mk

rpm: build/$(RPM_FILE)

pkg:
	VERSION=$(VERSION) RELEASE=$(RELEASE) ./scripts/pkg.sh

build/$(RPM_FILE): build/_$(APP_NAME)/Makefile $(TARGETS)
	cd build/_$(APP_NAME) && make package && cp -af $(RPM_FILE) ../
	rpm -pql build/$(RPM_FILE)

pkgs: $(TARGETS)
	APP_NAME=n9e-agentd make rpm
	sudo docker run --rm \
		-v $(PWD):/src \
		--name golang-cross-builder \
		--hostname golang-cross-builder \
		-e GO_BUILD_LDFLAGS='$(GO_BUILD_LDFLAGS)' \
		-e VERSION=$(VERSION) \
		-e RELEASE=$(RELEASE) \
		-it ghcr.io/gythialy/golang-cross-builder:v1.16.2 \
		/src/scripts/pkgs.sh

build/n9e-agentd: $(DEP_OBJS)
	unset GOFLAGS && \
	GO111MODULE=on CGO_ENABLED=$(CGO_ENABLED) \
	go build -ldflags '$(GO_BUILD_LDFLAGS)' \
	-o $@ ./cmd/agentd && \
	$@ version

build/agentdctl: $(DEP_OBJS)
	GO111MODULE=on CGO_ENABLED=$(CGO_ENABLED) \
	go build -ldflags '$(GO_BUILD_LDFLAGS)' \
	-o $@ ./cmd/agentdctl

clean:
	rm -rf build

release:
	VERSION=$(VERSION) RELEASE=$(RELEASE) ./scripts/release.sh

.PHONY: tools
tools:
	go get -u github.com/tcnksm/ghr
