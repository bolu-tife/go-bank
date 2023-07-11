build:
	@go build -o bin/go-bank

run: build
	@./bin/go-bank

test:
	@go test -v ./...

seed: build
	@./bin/go-bank -seed