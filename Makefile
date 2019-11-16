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

all: clean linux darwin test ## clean the previous output, run tests and generate linux, and darwin binaries

clean: ## clean the build
	-rm -f ${BINARY}-*
	-rm -f ${BINARY}


LDFLAGS=--ldflags="-X main.jtimonVersion=${COMMIT}-${BRANCH} -X main.buildTime=${TIME}"

linux: ## generate a linux version of the binary
	GOOS=linux GOARCH=${GOARCH} go build ${LDFLAGS} -o ${BINARY}-linux-${GOARCH} .

darwin: ## generate a osx version of the binary
	GOOS=darwin GOARCH=${GOARCH} go build ${LDFLAGS} -o ${BINARY}-darwin-${GOARCH} .

docker: ## build a docker image that can be used to execute the binary
	docker build --build-arg COMMIT=${COMMIT} --build-arg BRANCH=${BRANCH} --build-arg TIME=${TIME} -t jtimon . 
	ln -sf launch-docker-container.sh jtimon
	@echo "Usage: docker run -ti --rm jtimon --help"
	@echo "or simply call the shell script './jtimon --help"

docker-run:  ## run the docker image, output the jtimon help, ensure you previously have run make docker
	docker run -ti --rm -v ${PWD}:/root:ro jtimon --help || true

docker-sh: ## start the docker container and exec into shell, ensure you have previously run make docker 
	docker run -ti --rm -v ${PWD}:/root --entrypoint=/bin/sh jtimon

test: ## run the go tests
	go vet
	go test --covermode=count -v --coverprofile=coverage.out
	## go tool cover --html=coverage.out
	go tool cover --func=coverage.out

help: 
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: linux darwin docker docker-run docker-sh test help

