## How to generate test data


```
$ cd $GOPATH/src/github.com/nileshsimaria/jtimon
$ ./jtimon-darwin-amd64 --config tests/data/cisco-ios-xr/config/xr-all.json --generate-test-data
$ ./jtimon-darwin-amd64 --config tests/data/cisco-ios-xr/config/xr-wdsysmon.json --generate-test-data

$ ls tests/data/cisco-ios-xr/schema/
bgp.json                interfaces.json         native-statsd-oper.json wdsysmon.json

$ ls tests/data/cisco-ios-xr/config/
xr-all.json                xr-all.json.testexp        xr-wdsysmon.json           xr-wdsysmon.json.testexp   xr-wdsysmon.json.testres
xr-all.json.testbytes      xr-all.json.testmeta       xr-wdsysmon.json.testbytes xr-wdsysmon.json.testmeta
```

- .json is JTIMON config file
- .testmeta is produced with --generate-test-data
- .testbytes is proto encoded messages from router for testing. Its produced with --generate-test-data
- .testexp is produced when you run with --generate-test-data to hold expected results
- .testres is produced on "make test" (basically when we run the test)

There should be no diff between .testres and .testexp at all, otherwise test fails.


For Junos, use --no-per-packet-goroutines when generate test data.
```
$ ./jtimon-darwin-amd64 --config tests/data/juniper-junos/config/interfaces.json --no-per-packet-goroutines --generate-test-data
```
