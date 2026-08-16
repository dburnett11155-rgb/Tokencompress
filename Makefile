BINARY=tokencompress
VERSION=1.1.0

.PHONY: build test release clean

build:
	go build -ldflags="-s -w" -o $(BINARY) tokencompress.go

test:
	go test ./...

release: test build
	@echo "Release $(VERSION) built successfully"

clean:
	rm -f $(BINARY)
