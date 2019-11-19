# Copyright (c) 2018, Juniper Networks, Inc.
# All rights reserved.

FROM golang:1.13.4-alpine3.10 as builder
ARG COMMIT
ARG BRANCH
ARG TIME

WORKDIR /go/src/app
COPY . .

RUN GO111MODULE=on CGO_ENABLED=0 go build -mod vendor \
    --ldflags="-X main.jtimonVersion=${COMMIT}-${BRANCH} -X main.buildTime=${TIME}" \
    -o /usr/local/bin/jtimon

FROM alpine
COPY --from=builder /usr/local/bin/jtimon /usr/local/bin/jtimon

VOLUME /u
WORKDIR /u
ENTRYPOINT ["/usr/local/bin/jtimon"]
