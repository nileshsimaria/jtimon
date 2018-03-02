# Copyright (c) 2018, Juniper Networks, Inc.
# All rights reserved.

FROM alpine:3.7
RUN apk add --no-cache ethtool
RUN apk add --no-cache --virtual build-dependencies git go musl-dev \
  && go get github.com/golang/protobuf/proto \
  && go get github.com/gorilla/mux \
  && go get github.com/influxdata/influxdb/client/v2 \
  && go get github.com/prometheus/client_golang/prometheus/promhttp \
  && go get github.com/spf13/pflag \
  && go get golang.org/x/net/context \
  && go get google.golang.org/grpc \
  && go get github.com/nileshsimaria/jtimon/telemetry \
  && go get github.com/nileshsimaria/jtimon/authentication \
  && git clone https://github.com/nileshsimaria/jtimon.git \
  && cd jtimon && go build && strip jtimon && mv jtimon /usr/local/bin \
  && cd .. && rm -rf jtimon && apk del build-dependencies

VOLUME /u
WORKDIR /u
ENTRYPOINT ["/usr/local/bin/jtimon"]
