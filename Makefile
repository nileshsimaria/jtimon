# Copyright (c) 2018, Juniper Networks, Inc.
# All rights reserved.

all: run

build: 
	docker build -t jtimon .
	ln -sf launch-docker-container.sh jtimon
	@echo "Usage: docker run -ti --rm jtimon --help"
	@echo "or simply call the shell script './jtimon --help"

run: build
	docker run -ti --rm -v ${PWD}:/root:ro jtimon --help || true

sh: 
	docker run -ti --rm -v ${PWD}:/root --entrypoint=/bin/sh jtimon
