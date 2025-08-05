.PHONY: build
build: 
	env GOOS=linux GOARCH=arm64 go build -o ./bin/display-driver .