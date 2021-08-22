module github.com/hyperledger/fabric-droplib/component/listener

go 1.16

replace (
	github.com/hyperledger/fabric-droplib/base/log => ../../base/log
	github.com/hyperledger/fabric-droplib/base/services => ../../base/services
	github.com/hyperledger/fabric-droplib/component/base => ../base
	github.com/hyperledger/fabric-droplib/component/pubsub => ../pubsub
)

require (
	github.com/hyperledger/fabric-droplib/base/log v0.0.0-00010101000000-000000000000
	github.com/hyperledger/fabric-droplib/base/services v0.0.0-00010101000000-000000000000
	github.com/hyperledger/fabric-droplib/component/base v0.0.0-00010101000000-000000000000
	github.com/hyperledger/fabric-droplib/component/pubsub v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.7.0
)
