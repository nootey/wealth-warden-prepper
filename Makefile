BIN := bin/prepper

.PHONY: build test lint run clean release

build:
	go build -o $(BIN) ./cmd/prepper

test:
	go test ./...

lint:
	gofmt -l . && go vet ./...

run:
	go run ./cmd/prepper $(ARGS)

release:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/prepper-linux-amd64 ./cmd/prepper
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o dist/prepper-linux-arm64 ./cmd/prepper

clean:
	rm -rf bin dist
