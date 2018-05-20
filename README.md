# jtimon
Junos Telemetry Interface client

# Setup
<pre>
$ go get github.com/nileshsimaria/jtimon
$ $GOPATH/bin/jtimon --help

OR

$ git clone https://github.com/nileshsimaria/jtimon.git
$ cd jtimon
$ go build
$ ./jtimon --help
</pre>

# Docker container

Alternatively to building jtimon native, one can build a jtimon Docker container and
run it dockerized while passing the local directory to the container to access the 
json file.

To build the container:

<pre>
$ make build
</pre>

Check the resulting image:

<pre>
$ docker images jtimon
REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
jtimon              latest              3b7622e1464f        6 minutes ago       174MB
</pre>

Run it:

<pre>
$ docker run -ti --rm -v ${PWD}:/u:ro jtimon --help
</pre>

Or simply by calling ./jtimon, which is a symlink to launch-docker-container.sh, capable of launching the container by name with the current directory mounted into /u:

<pre>
$ ./jtimon
Enter config file name: bla.json
2018/03/02 13:53:44 File error: open bla.json: no such file or directory
</pre>

# CLI Options

<pre>
$ ./jtimon --help
Usage of ./jtimon:
      --api                     Receive HTTP commands when running
      --compression string      Enable HTTP/2 compression (gzip, deflate)
      --config strings          Config file name(s)
      --explore-config          Explore full config of JTIMON and exit
      --gnmi                    Use gnmi proto
      --gnmi-encoding string    gnmi encoding (proto | json | bytes | ascii | ietf-json (default "proto")
      --gnmi-mode string        Mode of gnmi (stream | once | poll (default "stream")
      --grpc-headers            Add grpc headers in DB
      --gtrace                  Collect GRPC traces
      --latency-profile         Profile latencies. Place them in TSDB
      --log-mux-stdout          All logs to stdout
      --max-run int             Max run time in seconds
      --pprof                   Profile JTIMON
      --pprof-port int32        Profile port (default 6060)
      --prefix-check            Report missing __prefix__ in telemetry packet
      --print                   Print Telemetry data
      --prometheus              Stats for prometheus monitoring system
      --prometheus-port int32   Prometheus port (default 8090)
      --stats-handler           Use GRPC statshandler
      --version                 Print version and build-time of the binary and exit
pflag: help requested

# Config
<pre>

To explore what can go in config, please use --explore-config option. Except connection details like host, port, etc no other part of the config is mandatory e.g. do not use influx in your config if you dont want to insert data into it. 

$ ./jtimon --explore-config
Version: version-not-available BuildTime build-time-not-available

{
    "host": "",
    "port": 0,
    "user": "",
    "password": "",
    "meta": false,
    "eos": false,
    "cid": "",
    "api": {
        "port": 0
    },
    "grpc": {
        "ws": 0
    },
    "tls": {
        "clientcrt": "",
        "clientkey": "",
        "ca": "",
        "servername": ""
    },
    "influx": {
        "server": "",
        "port": 0,
        "dbname": "",
        "user": "",
        "password": "",
        "recreate": false,
        "measurement": "",
        "diet": false,
        "batchsize": 0,
        "batchfrequency": 0,
        "retention-policy": ""
    },
    "paths": [
        {
            "path": "",
            "freq": 0,
            "mode": ""
        }
    ],
    "log": {
        "file": "",
        "verbose": false,
        "periodic-stats": 0,
        "drop-check": false,
        "latency-check": false,
        "csv-stats": false,
        "FileHandle": null,
        "Logger": null
    }
}
</pre>

