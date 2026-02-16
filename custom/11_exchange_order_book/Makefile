.PHONY: build execute

build:
	@mkdir -p bin
	@go build -o bin/orderbook .

execute: build
	@./bin/orderbook