# Copyright (c) 2018, Juniper Networks, Inc.
# All rights reserved.

all: run

LDFLAGS=--ldflags="-X main.Version=`git rev-list -1 HEAD` -X main.BuildTime=`date +%FT%T%z`"
build-mac:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o jtimon-darwin-amd64
build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o jtimon-linux-amd64

build: 
	docker build -t jtimon .
	ln -sf launch-docker-container.sh jtimon
	@echo "Usage: docker run -ti --rm jtimon --help"
	@echo "or simply call the shell script './jtimon --help"

run: build
	docker run -ti --rm -v ${PWD}:/root:ro jtimon --help || true

sh: 
	docker run -ti --rm -v ${PWD}:/root --entrypoint=/bin/sh jtimon
