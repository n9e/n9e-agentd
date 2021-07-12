build/_$(APP_NAME)/Makefile: $(TARGETS) Makefile
	mkdir -p build/_$(APP_NAME); cd build/_$(APP_NAME); cmake ../.. -DAPP_NAME=$(APP_NAME) -DCPACK_PACKAGE_VERSION="$(VERSION)" -DCPACK_PACKAGE_RELEASE="$(RELEASE)" -DCPACK_GENERATOR="$(GENERATOR)"

directories: build

build:
	mkdir -p build

devrun: build/n9e-agentd
	@echo "./build/n9e-agentd start --config ./run/etc/agentd.yml -v 10 --add-dir-header 2>&1"

dev: build/n9e-agentd
	CGO_ENABLED=$(CGO_ENABLED) APP_NAME=n9e-agentd watcher -logtostderr \
		 --add-dir-header -v 10 -e build -e .git -e docs \
		 -e plugins -e tmp -e vendor -e staging -f .go -d 1000

pkg/data/resources.go: pkg/data/resources/*
	go-bindata --prefix pkg/data/resources -pkg data -o $@ pkg/data/resources/...
