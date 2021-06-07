.PHONY: clean test

APP_NAME=agent
#DEP_OBJS=$(shell find . -name "*.go" -type f -not -path "./vendor/*" -and -not -path "./staging/*")
DEP_OBJS=$(shell find . -name "*.go" -type f -not -path "./vendor/*" -and -not -path "./staging/*")
TARGETS=directories build/$(APP_NAME)

GO_BUILD_LDFLAGS_CMD      := $(abspath ./scripts/go-build-ldflags.sh)
GO_BUILD_LDFLAGS          := $(shell $(GO_BUILD_LDFLAGS_CMD) LDFLAG)
GO_TAGS                   := jmx,kubelet

all: $(TARGETS)

.PHONY: devrun dev
	# /opt/datadog-agent/bin/agent/agent run -p /opt/datadog-agent/run/agent.pid
devrun:
	@echo "./build/agent -c ./etc/agent.yml"

.PHONY: run
run:
	./build/agent -c ./etc/agent.yml --vmodule=prometheus=10  2>&1

dev:
	APP_NAME=$(APP_NAME) watcher --logtostderr -v 10 -e build -e .git -e docs -e plugins -e tmp -e vendor -e staging -f .go -d 1000


build/$(APP_NAME): $(DEP_OBJS)
	GO111MODULE=on \
	go build -ldflags '$(GO_BUILD_LDFLAGS)' -tags '$(GO_TAGS)' \
	-o $@  ./cmd/$(APP_NAME) && \
	$@ version

directories: build

build:
	mkdir -p build

%.pb.go: %.proto
	protoc -I/usr/local/include -I. --go_out=plugins=grpc:$(GOPATH)/src $<

%.pb.gw.go: %.proto
	protoc -I/usr/local/include -I. \
	  -I$(GOPATH)/src \
	  -I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
	  --grpc-gateway_out=logtostderr=true:$(GOPATH)/src \
	  $<

clean:
	rm -rf build

.PHONY: env
env:
	yum install -y systemd-devel


.PHONY: mocker
mocker:
	go build -o build/mocker ./cmd/mocker && \
	./build/mocker

build/mocker: $(shell find cmd/mocker)
	go build -o $@  ./cmd/mocker

