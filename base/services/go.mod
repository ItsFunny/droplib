module github.com/hyperledger/fabric-droplib/base/services

go 1.16

replace github.com/hyperledger/fabric-droplib/base/log => ../log

require (
	github.com/hyperledger/fabric-droplib/base/log v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.5.1
)
