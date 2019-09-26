VERSION := $(shell cat version.txt)

.PHONY: all
all: vet lint release

.PHONY: vet
vet:
	go vet ./{cmd,pkg}/...

.PHONY: lint
lint:
	golint ./{cmd,pkg}/...

.PHONY: build-linux
build-linux:
	GOOS=linux go build -o build/linux/fastly-lem ./cmd/fastly-lem

.PHONY: build-windows
build-windows:
	GOOS=windows go build -o build/windows/fastly-lem ./cmd/fastly-lem

.PHONY: build-macos
build-macos:
	GOOS=darwin go build -o build/macos/fastly-lem ./cmd/fastly-lem

.PHONY: build
build: build-linux build-windows build-macos

.PHONY: clean
clean: 
	rm -rfv build/*

.PHONY: release
release: clean build 
	mkdir -p build/linux build/windows build/macos
	@echo Building version $(VERSION)
	tar cvfz build/fastly-lem-$(VERSION).tgz build/linux build/windows build/macos config/*
