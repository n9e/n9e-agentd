.PHONY: pkg pkgs var clean release tools

.EXPORT_ALL_VARIABLES:
VERSION=5.1.0
RELEASE=alpha.0
GENERATOR=RPM
APP_NAME?=n9e-agentd
CGO_ENABLED?=1
OBJ=$(APP_NAME)
RPM_FILE=$(APP_NAME)-$(VERSION)-$(RELEASE).$(shell uname -s).$(shell uname -m).rpm
DEP_OBJS=$(shell find pkg -name "*.go" -type f) \
	 pkg/data/resources.go
TARGETS?=build/n9e-agentd
GO_BUILD_LDFLAGS=`./scripts/go-build-ldflags.sh LDFLAG`


all: $(TARGETS)

rpm: build/$(RPM_FILE)

pkg: ./build/envs
	. ./build/envs && ./scripts/build.sh pkg

pkgs:
	./scripts/pkgs.sh

var:
	./scripts/go-build-envs.sh

clean:
	sudo rm -rf build/*

release:
	VERSION=${VERSION} RELEASE=${RELEASE} ./scripts/release.sh

tools:
	go get -u github.com/tcnksm/ghr

agent: ./build/envs
	. ./build/envs && ./scripts/build.sh

build/n9e-agentd: $(DEP_OBJS) ./build/envs
	. ./build/envs && ./scripts/build.sh

build/${RPM_FILE}: build/_${APP_NAME}/Makefile ${TARGETS}
	cd build/_${APP_NAME} && make package && cp -af ${RPM_FILE} ../
	rpm -pql build/$(RPM_FILE)

build/_${APP_NAME}/Makefile: ${TARGETS} Makefile
	./scripts/go-build-envs.sh cmake > build/envs.cmake && \
		mkdir -p build/_${APP_NAME} && \
		cd build/_${APP_NAME} && \
		cmake ../.. 

build/envs: Makefile
	./scripts/go-build-envs.sh export > $@

build/envs.cmake: Makefile
	./scripts/go-build-envs.sh cmake >  $@

pkg/data/resources.go:
	go-bindata --prefix pkg/data/resources -pkg data -o $@ pkg/data/resources/...

.PHONY: mocker
mocker:
	go build -o ./build/mocker ./cmd/mocker
