.PHONY: pkg pkgs var clean release dev devrun tools

.EXPORT_ALL_VARIABLES:
VERSION=5.1.0
RELEASE=rc1
GENERATOR=RPM
APP_NAME?=n9e-agentd
CGO_ENABLED?=1
OBJ=$(APP_NAME)
RPM_FILE=$(APP_NAME)-$(VERSION)-$(RELEASE).$(shell uname -s).$(shell uname -m).rpm
DEP_OBJS=$(shell find . -name "*.go" -type f -not -path "./vendor/*" -a -not -path "./staging/*") \
	 pkg/data/resources.go
TARGETS?=build/n9e-agentd
GO_BUILD_LDFLAGS=`./scripts/go-build-ldflags.sh LDFLAG`


all: $(TARGETS)

rpm: build/$(RPM_FILE)

pkg: ./build/envs
	source ./build/envs && ./scripts/build.sh pkg

pkgs: $(TARGETS)
	./scripts/pkgs.sh

var:
	./scripts/go-build-envs.sh

clean:
	sudo rm -rf build/*

release:
	VERSION=${VERSION} RELEASE=${RELEASE} ./scripts/release.sh

tools:
	go get -u github.com/tcnksm/ghr

build/n9e-agentd: $(DEP_OBJS) ./build/envs
	source ./build/envs && ./scripts/build.sh

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

devrun: build/n9e-agentd
	@echo "./build/n9e-agentd start --config ./run/etc/agentd.yml -v 10 --add-dir-header 2>&1"

dev: build/n9e-agentd
	source ./build/envs && watcher -logtostderr \
		 --add-dir-header -v 10 -e build -e .git -e docs \
		 -e plugins -e tmp -e vendor -e staging -f .go -d 1000

pkg/data/resources.go:
	go-bindata --prefix pkg/data/resources -pkg data -o $@ pkg/data/resources/...
