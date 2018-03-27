#cd to $GOPATH/src/github.com/nileshsimaria/jtimon/gnmi

protoc -I . --go_out=plugins=grpc,Mgoogle/protobuf/descriptor.proto=github.com/golang/protobuf/protoc-gen-go/descriptor:. gnmi.proto 
