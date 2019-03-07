APP_NAME=hermesctl

export GOOS=linux
export GOARCH=amd64

build: fmt
	GOOS=linux CGO_ENABLED=0 go build -o ../bin/$(APP_NAME)
	GOOS=darwin CGO_ENABLED=0 go build -o ../bin/$(APP_NAME)_darwin
	GOOS=windows CGO_ENABLED=0 go build -o ../bin/$(APP_NAME).exe

fmt:
	go fmt
