all: lint test build

install: lint test build
	sudo cp -af i3gopher /usr/local/bin/

lint:
	golangci-lint run

test:
	go test ./...

build:
	go build
