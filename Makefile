.PHONY: run test build clean

run:
	go run .

test:
	go vet ./...
	go test ./...

# Self-contained binary: config, templates and static files packed in.
build:
	go run ./hack/pack
	go build -tags packed -o bin/alpaca-app .

clean:
	rm -rf bin internal/packed
