module github.com/hyperledger/fabric-droplib/component/event

go 1.16

replace (
	github.com/hyperledger/fabric-droplib/base/log => ../../base/log
	github.com/hyperledger/fabric-droplib/base/services => ../../base/services
	github.com/hyperledger/fabric-droplib/component/base => ../base
)

require (
	github.com/hyperledger/fabric-droplib/base/log v0.0.0-00010101000000-000000000000
	github.com/hyperledger/fabric-droplib/base/services v0.0.0-00010101000000-000000000000
	github.com/hyperledger/fabric-droplib/component/base v0.0.0-00010101000000-000000000000
	github.com/libp2p/go-libp2p-core v0.8.5
	github.com/stretchr/testify v1.7.0
)
