.PHONY: all
all: test build

.PHONY: install
install: lint test build
	sudo cp -af i3gopher /usr/local/bin/

.PHONY: lint
lint:
	make -C gotools golangci-lint
	./gotools/golangci-lint run

.PHONY: test
test:
	go test ./...

.PHONY: build
build:
	go build
