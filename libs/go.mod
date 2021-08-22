module github.com/hyperledger/fabric-droplib/libs

go 1.16

replace (
	github.com/hyperledger/fabric-droplib/base/log => ../base/log
	github.com/hyperledger/fabric-droplib/base/services => ../base/services
	github.com/hyperledger/fabric-droplib/common => ../common
	github.com/hyperledger/fabric-droplib/structure => ../structure
)

require (
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.5.2
	github.com/hyperledger/fabric-droplib/base/log v0.0.0-00010101000000-000000000000
	github.com/hyperledger/fabric-droplib/base/services v0.0.0-00010101000000-000000000000
	github.com/hyperledger/fabric-droplib/common v0.0.0-00010101000000-000000000000
	github.com/hyperledger/fabric-droplib/structure v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.5.1
)
