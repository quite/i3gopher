.PHONY: all
all: test build

.PHONY: install
install: lint test build
	sudo cp -af i3gopher /usr/local/bin/

.PHONY: lint
lint: golangci-lint
	./golangci-lint run

golangci-lint: go.mod go.sum
	go build github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: test
test:
	go test ./...

.PHONY: build
build:
	go build
