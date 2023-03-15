# JTI Protobuf Schemas

These protobuf schemas were downloaded from https://webdownload.juniper.net/swdl/dl/secure/site/1/record/77110.html
In each schema, an option has been added so that all generated go is in the same package:
```
option go_package = "git.juniper.net/iceberg/healthbot-ingest-jti-native/pkg/jti/adapters/_junos/schema";
```

Also, to avoid protoc warnings about syntax not being specified, inline_jflow.proto, optics.proto, pbj.proto and qmon.proto have all had a line added to set syntax to proto2.

Otherwise the schemas are unchanged.

## Generating go Source

protoc and the gogo plugin must be installed. 
See https://github.com/golang/protobuf#installation && https://github.com/gogo/protobuf#getting-started.

```sh
$ protoc -I . --gogofast_out=Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor,plugins=grpc:../../../../../../../.. *proto
```
This will generate a .pb.go corresponding to each schema in the same directory.
The ../../../../../../../.. references the GO root directory (i.e. the directory containing git.juniper.net).


#### Note
Currently the healthbot-ingest-jti-native uses `Junos 21.4R1.12`  proto files
