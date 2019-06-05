PKG:=github.com/sapcc/hermes-ctl
APP_NAME:=hermesctl
PWD:=$(shell pwd)
UID:=$(shell id -u)
VERSION:=$(shell git describe --tags --always --dirty="-dev")
LDFLAGS:=-X $(PKG)/client.Version=$(VERSION)

export GO111MODULE:=off
export GOPATH:=$(PWD):$(PWD)/gopath
export CGO_ENABLED:=0

build: gopath/src/$(PKG) fmt
	GOOS=linux go build -ldflags="$(LDFLAGS)" -o bin/$(APP_NAME)
	GOOS=darwin go build -ldflags="$(LDFLAGS)" -o bin/$(APP_NAME)_darwin
	GOOS=windows go build -ldflags="$(LDFLAGS)" -o bin/$(APP_NAME).exe

docker:
	docker run -ti --rm -e GOCACHE=/tmp -v $(PWD):/$(APP_NAME) -u $(UID):$(UID) --workdir /$(APP_NAME) golang:latest make

fmt:
	gofmt -s -w acceptance audit client *.go

gopath/src/$(PKG):
	mkdir -p gopath/src/$(shell dirname $(PKG))
	ln -sf ../../../.. gopath/src/$(PKG)
