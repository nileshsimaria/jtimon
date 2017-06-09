# jtimon
Junos Telemetry Interface client

# Setup
<pre>
1. Install golang
2. Set GOROOT/GOPATH as needed
3. git clone https://github.com/nileshsimaria/jtimon.git
4. Fetch following dependent packages

    $ go get github.com/golang/protobuf/proto
    $ go get github.com/gorilla/mux
    $ go get github.com/influxdata/influxdb/client/v2
    $ go get github.com/prometheus/client_golang/prometheus/promhttp
    $ go get github.com/spf13/pflag
    $ go get golang.org/x/net/context
    $ go get google.golang.org/grpc

5. cd jtimon
6. go build
7. ./jtimon --help
</pre>

# CLI Options

<pre>
$ ./jtimon --help
Usage of ./jtimon:
      --cert string                   CA certificate file
      --compression string            Enable HTTP/2 compression (gzip, deflate)
      --config string                 Config file name
      --drop-check                    Check for packet drops
      --gtrace                        Collect GRPC traces
      --latency-check                 Check for latency
      --log string                    Log file name
      --max-kv int                    Max kv
      --max-run int                   Max run time in seconds
      --pdt                           PDT style influx DB schema
      --prefix-check                  Report missing __prefix__ in telemetry packet
      --print                         Print Telemetry data
      --prometheus                    Stats for prometheus monitoring system
      --server-host-override string   ServerName used to verify the hostname
      --sleep int                     Sleep after each read (ms)
      --stats int                     Collect and Print statistics periodically
      --time-diff                     Time Diff for sensor analysis using InfluxDB
      --tls                           Connection uses TLS
</pre>      

# Config
<pre>
Sample JSON config file to subscribe /interfaces @2s, /bpg @10s and /components @10s.

{
    "host": "Junos-Device-IP",
    "port": 50051, ## Junos gRPC port
    "user": "uname", 
    "password": "pwd",
    "cid": "cid-2", ## unique client ID
    "grpc" : {
        "ws" : 524288 ## advertise HTTP2 window size (512K) (default 64K)
    },
    "api": {
        "port" : 7878 ## send pause / unpause command to this port (optional)
    },
    "influx" : { ## influx DB config (optional)
        "server" : "127.0.0.1",
        "port" : 8086,
        "dbname" : "vptx-db",
        "measurement" : "vptx",
        "recreate" : true, ## Recreate the said DB (nuke old one)
        "user" : "influx",
        "password" : "influxdb"
    },    
    "paths": [{
        "path": "/interfaces",
        "freq": 2000
	}, {
        "path": "/bgp",
        "freq": 10000
	}, {
        "path": "/components",
        "freq": 10000
    }]
}

Sample run which collect the stats too.
$ ./jtimon --config <your-json-file> --stats <n> 
</pre>
