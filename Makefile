.EXPORT_ALL_VARIABLES:
PACKAGE_VERSION=5.1.0
PACKAGE_RELEASE=alpha1
DEP_OBJS=$(shell find pkg -name "*.go" -type f) pkg/data/resources.go
TARGETS?=build/agentd

all: $(TARGETS)

.PHONY: pkg
pkg:
	./scripts/pkg.sh

.PHONY: pkgs
pkgs:
	./scripts/pkgs.sh

.PHONY: omnibus
omnibus:
	./scripts/omnibus.sh

.PHONY: clean
clean:
	rm -rf build/*

.PHONY: release
release:
	./scripts/release.sh

.PHONY: tools
tools:
	go get -u github.com/tcnksm/ghr

.PHONY: mocker
mocker:
	go build -o ./build/mocker ./cmd/mocker

build/agentd: $(DEP_OBJS)
	./scripts/build.sh

pkg/data/resources.go:
	go-bindata --prefix pkg/data/resources -pkg data -o $@ pkg/data/resources/...
