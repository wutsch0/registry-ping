BINARY   := registry-ping
CONFIG   ?= config.yaml

.PHONY: build run test lint clean

build:
	go build -o $(BINARY) ./cmd/registry-ping/

run: build
	./$(BINARY) -config $(CONFIG)

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -f $(BINARY)
