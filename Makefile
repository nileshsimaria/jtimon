# Copyright (c) 2018, Juniper Networks, Inc.
# All rights reserved.

BINARY = jtimon
GOARCH = amd64

COMMIT=$(shell git rev-parse HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
TIME=$(shell date +%FT%T%z)

GITHUB_USERNAME=nileshsimaria
BUILD_DIR=${GOPATH}/src/github.com/${GITHUB_USERNAME}/${BINARY}
CURRENT_DIR=$(shell pwd)
BUILD_DIR_LINK=$(shell readlink ${BUILD_DIR})

all: clean linux darwin windows

clean:
	-rm -f ${BINARY}-*
	-rm -f ${BINARY}


LDFLAGS=--ldflags="-X main.version=${COMMIT}-${BRANCH} -X main.buildTime=${TIME}"

linux: 
	GOOS=linux GOARCH=${GOARCH} go build ${LDFLAGS} -o ${BINARY}-linux-${GOARCH} .

darwin:
	GOOS=darwin GOARCH=${GOARCH} go build ${LDFLAGS} -o ${BINARY}-darwin-${GOARCH} .

windows:
	GOOS=windows GOARCH=${GOARCH} go build ${LDFLAGS} -o ${BINARY}-windows-${GOARCH}.exe .

docker: 
	docker build -t jtimon .
	ln -sf launch-docker-container.sh jtimon
	@echo "Usage: docker run -ti --rm jtimon --help"
	@echo "or simply call the shell script './jtimon --help"

docker-run: build
	docker run -ti --rm -v ${PWD}:/root:ro jtimon --help || true

docker-sh: 
	docker run -ti --rm -v ${PWD}:/root --entrypoint=/bin/sh jtimon

.PHONY: linux darwin windows docker docker-run docker-sh

