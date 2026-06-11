.PHONY: test build run

test:
	go test ./server/tests/

build:
	go build -o mtg-alternatives .

run: build
	./mtg-alternatives
