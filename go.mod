module github.com/nileshsimaria/jtimon

go 1.17

require (
	github.com/Shopify/sarama v1.26.1
	github.com/cenkalti/backoff/v4 v4.1.3
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.5.2
	github.com/influxdata/influxdb v1.5.2
	github.com/jhump/protoreflect v1.15.1
	github.com/nileshsimaria/jtimon/schema v0.0.0-00010101000000-000000000000
	github.com/nileshsimaria/jtisim v0.0.0-20190103005352-93e1756cba8f
	github.com/nileshsimaria/lpserver v0.0.0-20201223125141-2c2ca16a8d88
	github.com/openconfig/grpctunnel v0.0.0-20220412182401-2d6d258f73d1
	github.com/prometheus/client_golang v0.8.0
	github.com/spf13/pflag v1.0.1
	golang.org/x/net v0.7.0
	google.golang.org/grpc v1.53.0
	google.golang.org/protobuf v1.28.2-0.20230222093303-bc1253ad3743
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bufbuild/protocompile v0.4.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/eapache/go-resiliency v1.2.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/golang/snappy v0.0.3 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/jcmturner/gofork v1.0.0 // indirect
	github.com/klauspost/compress v1.9.8 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/pierrec/lz4 v2.4.1+incompatible // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.0.0-20180413074202-d0f7cd64bda4 // indirect
	github.com/prometheus/procfs v0.0.0-20180408092902-8b1c2da0d56d // indirect
	github.com/rcrowley/go-metrics v0.0.0-20190826022208-cac0b30c2563 // indirect
	golang.org/x/crypto v0.0.0-20211108221036-ceb1ce70b4fa // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	google.golang.org/genproto v0.0.0-20230110181048-76db0878b65f // indirect
	gopkg.in/jcmturner/aescts.v1 v1.0.1 // indirect
	gopkg.in/jcmturner/dnsutils.v1 v1.0.1 // indirect
	gopkg.in/jcmturner/gokrb5.v7 v7.5.0 // indirect
	gopkg.in/jcmturner/rpc.v1 v1.1.0 // indirect
)

replace github.com/nileshsimaria/jtimon/schema => ./schema

replace github.com/nileshsimaria/jtimon/schema/telemetry_top => ./schema/telemetry_top
