.PHONY: build test install clean

BINARY_NAME=tuido
BIN_DIR=bin

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY_NAME) .

test:
	go test ./...

install: build
	sudo install -m 0755 $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)

clean:
	rm -f $(BIN_DIR)
