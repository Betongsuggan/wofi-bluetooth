.PHONY: build install clean

BINARY_NAME=wofi-bluetooth
INSTALL_PATH=/usr/local/bin

build:
	go build -o $(BINARY_NAME) ./cmd/wofi-bluetooth

install: build
	install -m 755 $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)
	go clean

test:
	go test -v ./...
