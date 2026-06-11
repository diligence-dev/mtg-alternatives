.PHONY: test build run

test:
	go test -v ./server/...

build:
	go build -o mtg-alternatives .

run: build
	./mtg-alternatives
