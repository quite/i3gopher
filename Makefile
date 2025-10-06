.PHONY: all
all: test build

.PHONY: install
install: all
	sudo install i3gopher /usr/local/bin/

.PHONY: test
test:
	go test ./...

.PHONY: build
build:
	go build

golangci_version=v1.61.0
golangci_cachedir=$(HOME)/.cache/golangci-lint/$(golangci_version)
.PHONY: lint
lint:
	mkdir -p $(golangci_cachedir)
	podman run --rm -it \
		-v $$(pwd):/src -w /src \
		-v $(golangci_cachedir):/root/.cache \
		docker.io/golangci/golangci-lint:$(golangci_version)-alpine \
		golangci-lint run
