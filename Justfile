set dotenv-load := true

build:
    go build -ldflags="-s -w" -o importer

run command:
    go run main.go -- {{command}}