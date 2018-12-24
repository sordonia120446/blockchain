all: start

build:
	@echo "To build, be sure to set GOPATH to the working directory."
	@echo "Then run 'go get' to install {spew, mux, godotenv}"

help:
	@echo "Just run make bruh"

post:
	@curl -X POST -H "Content-Type: text/plain" --data '{"BPM": 150}' http://localhost:8080/
	@echo "\n"

start:
	@go run main.go

.PHONY: help, post, start
