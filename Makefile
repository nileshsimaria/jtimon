# Copyright (c) 2018, Juniper Networks, Inc.
# All rights reserved.

all: run

build: 
	docker build -t jtimon .

run: build
	docker run -ti --rm -v ${PWD}:/root jtimon --help

sh: 
	docker run -ti --rm -v ${PWD}:/root --entrypoint=/bin/sh jtimon
