module github.com/hyperledger/fabric-droplib/plugin/watcher

go 1.16

replace (
	github.com/hyperledger/fabric-droplib/base/log => ../../base/log
	github.com/hyperledger/fabric-droplib/base/services => ../../base/services
	github.com/hyperledger/fabric-droplib/common => ../../common
	github.com/hyperledger/fabric-droplib/component/base => ../../component/base
	github.com/hyperledger/fabric-droplib/component/listener => ../../component/listener
	github.com/hyperledger/fabric-droplib/component/pubsub => ../../component/pubsub
	github.com/hyperledger/fabric-droplib/component/routine => ../../component/routine
	github.com/hyperledger/fabric-droplib/libs => ../../libs
	github.com/hyperledger/fabric-droplib/structure => ../../structure
)

require (
	github.com/hyperledger/fabric-droplib/base/log v0.0.0-00010101000000-000000000000
	github.com/hyperledger/fabric-droplib/base/services v0.0.0-00010101000000-000000000000
	github.com/hyperledger/fabric-droplib/component/listener v0.0.0-00010101000000-000000000000
	github.com/hyperledger/fabric-droplib/component/routine v0.0.0-00010101000000-000000000000
	github.com/hyperledger/fabric-droplib/libs v0.0.0-00010101000000-000000000000
	github.com/hyperledger/fabric-droplib/structure v0.0.0-00010101000000-000000000000
	github.com/sasha-s/go-deadlock v0.3.1
	github.com/stretchr/testify v1.7.0
)
