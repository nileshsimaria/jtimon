[![Build Status](https://travis-ci.org/nileshsimaria/jtisim.svg?branch=master)](https://travis-ci.org/nileshsimaria/jtisim)

# jtisim
JTI simulator (server)

A stand alone program to stream JTI telemetry data. Useful in scale testing to produce high volume telemetry data.

<pre>
$ git clone https://github.com/nileshsimaria/jtisim.git
$ cd $GOPATH/src/github.com/nileshsimaria/jtisim/cmd
$ go build

$ ./cmd --help
Usage of ./cmd:
      --description-dir string   description directory (default "../desc")
      --host string              host name or ip (default "127.0.0.1")
      --port int32               grpc server port (default 50051)
      --random                   Use random number to generate counter values
</pre>
