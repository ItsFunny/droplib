module github.com/hyperledger/fabric-droplib/component/p2p

go 1.16


replace (
	github.com/hyperledger/fabric-droplib/base/log => ../../base/log
	github.com/hyperledger/fabric-droplib/base/services => ../../base/services
	github.com/hyperledger/fabric-droplib/component/base => ../base
)