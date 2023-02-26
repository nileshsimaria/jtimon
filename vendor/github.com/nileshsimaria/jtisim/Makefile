# Copyright (c) 2018, Juniper Networks, Inc.
# All rights reserved.

BINARY = jtisim
GOARCH = amd64

COMMIT=$(shell git rev-parse HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
TIME=$(shell date +%FT%T%z)
TAG=$(shell git describe --abbrev=0)

GITHUB_USERNAME=nileshsimaria
BUILD_DIR=${GOPATH}/src/github.com/${GITHUB_USERNAME}/${BINARY}
CURRENT_DIR=$(shell pwd)
BUILD_DIR_LINK=$(shell readlink ${BUILD_DIR})

all: clean linux darwin ## clean the previous output, run tests and generate linux, and darwin binaries

clean: ## clean the build
	-rm -f ${BINARY}-*
	-rm -f ${BINARY}


LDFLAGS=--ldflags="-X main.jtisimVersion=${TAG}-${COMMIT}-${BRANCH} -X main.buildTime=${TIME}"

linux: ## generate a linux version of the binary
	GOOS=linux GOARCH=${GOARCH} go build ${LDFLAGS} -o ${BINARY}-linux-${GOARCH} cmd/main.go

darwin: ## generate a osx version of the binary
	GOOS=darwin GOARCH=${GOARCH} go build ${LDFLAGS} -o ${BINARY}-darwin-${GOARCH} cmd/main.go

.PHONY: linux darwin

