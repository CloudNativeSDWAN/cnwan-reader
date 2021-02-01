module github.com/CloudNativeSDWAN/cnwan-reader

go 1.14

require (
	cloud.google.com/go v0.72.0
	github.com/CloudNativeSDWAN/cnwan-operator v0.3.0
	github.com/aws/aws-sdk-go v1.37.2
	github.com/google/go-cmp v0.5.4
	github.com/rs/zerolog v1.19.0
	github.com/spf13/cobra v1.0.0
	github.com/stretchr/testify v1.6.1
	go.etcd.io/etcd v0.0.0-20191023171146-3cf2f69b5738
	golang.org/x/oauth2 v0.0.0-20201208152858-08078c50e5b5
	google.golang.org/api v0.36.0
	google.golang.org/genproto v0.0.0-20201210142538-e3217bee35cc
	gopkg.in/yaml.v2 v2.3.0
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
	k8s.io/api v0.18.6
)

replace (
	github.com/coreos/etcd => github.com/coreos/etcd v3.3.25+incompatible
	go.etcd.io/bbolt => go.etcd.io/bbolt v1.3.5
	go.etcd.io/etcd => go.etcd.io/etcd v0.5.0-alpha.5.0.20200910180754-dd1b699fc489 // ae9734ed278b is the SHA for git tag v3.4.13
	google.golang.org/grpc => google.golang.org/grpc v1.27.1
)
