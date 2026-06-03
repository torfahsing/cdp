.PHONY: build install test test-integration clean release

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X github.com/user/cdp/cmd.Version=$(VERSION)
DIST := dist

build:
	go build -ldflags "$(LDFLAGS)" -o cdp .

install: build
	install -m 755 cdp $(HOME)/.local/bin/cdp

test:
	go test ./...

test-integration: build
	go test -tags integration ./test/ -v

clean:
	rm -f cdp
	rm -rf $(DIST)

release: clean
	mkdir -p $(DIST)
	GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(DIST)/cdp-darwin-arm64 .
	GOOS=darwin  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST)/cdp-darwin-amd64 .
	GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST)/cdp-linux-amd64 .
	GOOS=linux   GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(DIST)/cdp-linux-arm64 .
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST)/cdp-windows-amd64.exe .
	@echo "Binaries in $(DIST)/"
	@ls -lh $(DIST)/
