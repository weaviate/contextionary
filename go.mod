module github.com/semi-technologies/contextionary

require (
	github.com/fatih/camelcase v1.0.0
	github.com/golang/protobuf v1.4.2
	github.com/jessevdk/go-flags v1.4.0
	github.com/semi-technologies/weaviate v1.1.0
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1
	github.com/syndtr/goleveldb v0.0.0-20180708030551-c4c61651e9e3
	google.golang.org/grpc v1.24.0
)

replace github.com/coreos/go-systemd => github.com/coreos/go-systemd/v22 v22.0.0

go 1.13
