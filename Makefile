BIN_NAME = todo

run: build
	./bin/$(BIN_NAME)

build:
	go build -o ./bin/$(BIN_NAME) .
