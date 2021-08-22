module github.com/hyperledger/fabric-droplib/component/routine

go 1.16

replace (
	github.com/hyperledger/fabric-droplib/base/log => ../../base/log
	github.com/hyperledger/fabric-droplib/base/services => ../../base/services
	github.com/hyperledger/fabric-droplib/common => ../../common
	github.com/hyperledger/fabric-droplib/component/base => ../base
	github.com/hyperledger/fabric-droplib/libs => ../../libs
)

require (
	github.com/emirpasic/gods v1.12.0
	github.com/hyperledger/fabric-droplib/base/log v0.0.0-00010101000000-000000000000
	github.com/hyperledger/fabric-droplib/base/services v0.0.0-00010101000000-000000000000
	github.com/hyperledger/fabric-droplib/component/base v0.0.0-00010101000000-000000000000
	github.com/hyperledger/fabric-droplib/libs v0.0.0-00010101000000-000000000000
	github.com/panjf2000/ants/v2 v2.4.6
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	gopkg.in/eapache/queue.v1 v1.1.0
)
