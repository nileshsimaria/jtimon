[![Build Status](https://travis-ci.org/nileshsimaria/jtimon.svg?branch=master)](https://travis-ci.org/nileshsimaria/jtimon)

# jtimon

Junos Telemetry Interface client

## Setup

```sh
go get github.com/nileshsimaria/jtimon
$GOPATH/bin/jtimon --help
```

OR

```sh
git clone https://github.com/nileshsimaria/jtimon.git
cd jtimon
go build or make
./jtimon --help
```

Please note that if you use make to build source, it will produce binary with GOOS and GOARCH names e.g. jtimon-darwin-amd64, jtimon-linux-amd64 etc. Building the source using make is recommended as it will insert git-revision, build-time info in the binary.

To understand what targets are available in make, run the make help command as follows:

```sh
make help
```

### Note

If you are cloning the source and building it, please make sure you have environment variable GOPATH is set correctly.
https://golang.org/doc/code.html#GOPATH

## Docker container

Alternatively to building jtimon native, one can build a jtimon Docker container and run it dockerized while passing the local directory to the container to access the json file.

To build the container:

```sh
make docker
```

Check the resulting image:

```sh
$ docker images jtimon
REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
jtimon              latest              3b7622e1464f        6 minutes ago       174MB
```

Run it:

```sh
docker run -ti --rm -v ${PWD}:/u:ro jtimon --help
```

Or simply by calling ./jtimon, which is a symlink to launch-docker-container.sh, capable of launching the container by name with the current directory mounted into /u:

```sh
$ ./jtimon
Enter config file name: bla.json
2018/03/02 13:53:44 File error: open bla.json: no such file or directory
```

## CLI Options

```
$ ./jtimon-darwin-amd64 --help
Usage of ./jtimon-darwin-amd64:
      --compression string         Enable HTTP/2 compression (gzip)
      --config strings             Config file name(s)
      --config-file-list string    List of Config files
      --consume-test-data          Consume test data
      --explore-config             Explore full config of JTIMON and exit
      --generate-test-data         Generate test data
      --json                       Convert telemetry packet into JSON
      --log-mux-stdout             All logs to stdout
      --max-run int                Max run time in seconds
      --no-per-packet-goroutines   Spawn per packet go routines
      --pprof                      Profile JTIMON
      --pprof-port int32           Profile port (default 6060)
      --prefix-check               Report missing __prefix__ in telemetry packet
      --print                      Print Telemetry data
      --prometheus                 Stats for prometheus monitoring system
      --prometheus-port int32      Prometheus port (default 8090)
      --stats-handler              Use GRPC statshandler
      --version                    Print version and build-time of the binary and exit
```

## Config

To explore what can go in config, please use --explore-config option.

Except connection details like host, port, etc no other part of the config is mandatory e.g. do not use influx in your config if you dont want to insert data into it.

```
$ ./jtimon-darwin-amd64 --explore-config
2019/01/04 18:21:08 Version: e0ce6eccd0a02cc346bcd4e5e038d19b6747d33b-master BuildTime 2019-01-04T18:18:55-0800
2019/01/04 18:21:08
{
    "port": 0,
    "host": "",
    "user": "",
    "password": "",
    "cid": "",
    "meta": false,
    "eos": false,
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
        "batchsize": 0,
        "batchfrequency": 0,
        "retention-policy": "",
        "accumulator-frequency": 0
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
        "periodic-stats": 0,
        "verbose": false
    },
    "vendor": {
        "name": "",
        "remove-namespace": false,
        "schema": null
    },
    "alias": ""
}
```

I am explaining some config options which are not self-explanatory.

<pre>
meta : send username and password over gRPC meta instead of invoking LoginCheck() RPC for authentication. 
Please use SSL/TLS for security. For more details on how to use SSL/TLS, please refer wiki
https://github.com/nileshsimaria/jtimon/wiki/SSL
</pre>

<pre>
cid : client id. Junos expects unique client ids if multiple clients are subscribing to telemetry streams.
</pre>

<pre>
eos : end of sync. Tell Junos to send end of sync for on-change subscriptions.
</pre>

<pre>
grpc/ws : window size of grpc for slower clients
</pre>
