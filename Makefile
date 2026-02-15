.PHONY: build run test lint clean install uninstall

BINARY     := lanchr
BUILD_DIR  := ./bin
PREFIX     ?= /usr/local
VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS    := -s -w \
	-X github.com/zhengda-lu/lanchr/internal/cli.version=$(VERSION)

build:
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) ./cmd/lanchr

run: build
	$(BUILD_DIR)/$(BINARY)

test:
	go test -race -cover ./...

lint:
	golangci-lint run ./...

vet:
	go vet ./...

install: build
	install -d $(PREFIX)/bin
	install -m 755 $(BUILD_DIR)/$(BINARY) $(PREFIX)/bin/$(BINARY)

uninstall:
	rm -f $(PREFIX)/bin/$(BINARY)

clean:
	rm -rf $(BUILD_DIR)
