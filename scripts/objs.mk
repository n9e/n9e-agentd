build/_$(APP_NAME)/Makefile: $(TARGETS) Makefile
	mkdir -p build/_$(APP_NAME); cd build/_$(APP_NAME); cmake ../.. -DAPP_NAME=$(APP_NAME) -DCPACK_PACKAGE_VERSION="$(VERSION)" -DCPACK_PACKAGE_RELEASE="$(RELEASE)" -DCPACK_GENERATOR="$(GENERATOR)"

.PHONY: build/grabui
build/grabui:

build/grab: $(DEP_OBJS)
	GO111MODULE=on CGO_ENABLED=$(CGO_ENABLED) GOARCH=amd64 go build -ldflags "-w -s" -o $@ ./cmd/grab && $@ version

build/grabd: $(DEP_OBJS)
	GO111MODULE=on CGO_ENABLED=$(CGO_ENABLED) GOARCH=amd64 go build -ldflags "-w -s" -o $@ ./cmd/grabd && $@ version

cmd/$(APP_NAME)/env.go: $(shell find .git/refs -type f) Makefile
	echo "package main" > $@ \
		&& echo "const (" >> $@ \
		&& echo "VERSION =\"$(VERSION)\"" >> $@ \
		&& echo "RELEASE =\"$(RELEASE)\"" >> $@ \
		&& echo "COMMIT =\"$(shell git log -1 --pretty=%h)\"" >> $@ \
		&& echo "BUILDTIME = $(shell LANG=C date -u '+%s')" >> $@ \
		&& echo "BUILDER = \"$(shell whoami)@$(shell hostname)\"" >> $@ \
		&& echo "CHANGELOG =\`" >> $@ \
		&& git log --format='* %cd %aN%n- (%h) %s%d%n' --date=local | grep 'feature\|bugfix' | sed 's/[0-9]+:[0-9]+:[0-9]+ //' | sed -e 's/`/\"/g'>> $@ \
		&& echo "\`)" >> $@ \
		&& go fmt $@

pkg/data/resources.go: pkg/data/resources/*
	go-bindata --prefix pkg/data/resources -pkg data -o $@ pkg/data/resources/...

directories: build

build:
	mkdir -p build

ui/dist:
	cd ui && make build
