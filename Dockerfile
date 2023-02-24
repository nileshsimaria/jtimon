# Copyright (c) 2018, Juniper Networks, Inc.
# All rights reserved.

FROM golang:1.18-alpine as builder
ARG COMMIT
ARG BRANCH
ARG TIME
ARG TAG

WORKDIR /go/src/app
COPY . .

RUN GO111MODULE=on CGO_ENABLED=0 go build -mod vendor \
    --ldflags="-X main.jtimonVersion=${TAG}-${COMMIT}-${BRANCH} -X main.buildTime=${TIME}" \
    -o /usr/local/bin/jtimon && CGO_ENABLED=0 go test --covermode=count -v --coverprofile=coverage.out

FROM alpine
COPY --from=builder /usr/local/bin/jtimon /usr/local/bin/jtimon

VOLUME /u
WORKDIR /u
#RUN mkdir -p certs/self_signed/
#COPY ./certs/self_signed/ certs/self_signed
ENTRYPOINT ["/usr/local/bin/jtimon"]
