NAME = "server"

build:
	go build -o bin/${NAME} -ldflags="-s -w" .