module github.com/hyperledger/fabric-droplib

go 1.16

replace (
	github.com/Hyperledger-TWGC/tjfoc-gm => ../../../github.com/Hyperledger-TWGC/tjfoc-gm
	github.com/cloudflare/cfssl => /Users/joker/go/src/github.com/cloudflare/cfssl
	github.com/tjfoc/gmsm => ../../../github.com/tjfoc/gmsm
)

require (
	github.com/ChainSafe/go-schnorrkel v0.0.0-20210222182958-bd440c890782 // indirect
	github.com/cloudflare/cfssl v1.5.0
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.4.0
	github.com/google/btree v1.0.1 // indirect
	github.com/google/orderedcode v0.0.1
	github.com/hdevalence/ed25519consensus v0.0.0-20210204194344-59a8610d2b87 // indirect
	github.com/libp2p/go-libp2p v0.13.0
	github.com/libp2p/go-libp2p-connmgr v0.2.4
	github.com/libp2p/go-libp2p-core v0.8.0
	github.com/libp2p/go-libp2p-pubsub v0.4.1
	github.com/libp2p/go-libp2p-tls v0.1.3
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b
	github.com/multiformats/go-multiaddr v0.3.1
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.9.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.3.0
	github.com/stretchr/testify v1.6.1
	github.com/syndtr/goleveldb v1.0.0
	github.com/tjfoc/gmsm v0.0.0-00010101000000-000000000000
	github.com/x-cray/logrus-prefixed-formatter v0.5.2 // indirect
	go.uber.org/atomic v1.6.0
	golang.org/x/crypto v0.0.0-20201012173705-84dcc777aaee
	google.golang.org/protobuf v1.23.0
	gopkg.in/eapache/queue.v1 v1.1.0
)
