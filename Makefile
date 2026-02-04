.PHONY: build test install clean

BINARY_NAME=tuido

build:
	go build -o $(BINARY_NAME) ./cmd/tuido

test:
	go test ./...

install: build
	sudo install -m 0755 $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)
