# jtimon
Junos Telemetry Interface client

# Setup

1. Install golang
2. Set GOROOT/GOPATH as needed
3. git clone https://github.com/nileshsimaria/jtimon.git
4. Fetch following dependent packages

    go get github.com/golang/protobuf/proto<br>
    go get github.com/gorilla/mux<br>
    go get github.com/influxdata/influxdb/client/v2<br>
    go get github.com/prometheus/client_golang/prometheus/promhttp<br>
    go get github.com/spf13/pflag<br>
    go get golang.org/x/net/context<br>
    go get google.golang.org/grpc<br>

5. cd jtimon
6. go build
7. ./jtimon --help
