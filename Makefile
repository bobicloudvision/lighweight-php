.PHONY: build install clean test

build:
	go build -o lightweight-php .

install: build
	sudo cp lightweight-php /usr/local/bin/

clean:
	rm -f lightweight-php

test:
	go test ./...

deps:
	go mod download
	go mod tidy
