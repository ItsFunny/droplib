protoc --proto_path=/Users/joker/Desktop/gopath/src \
-I /Users/joker/Desktop/gopath/src/github.com/gogo/protobuf \
-I . --go_out=plugins=grpc,paths=source_relative:.  --go-grpc_out=require_unimplemented_servers=false,paths=source_relative:. ./types.proto

protoc --proto_path=/Users/joker/Desktop/gopath/src \
-I /Users/joker/Desktop/gopath/src/github.com/gogo/protobuf \
-I . --go_out=plugins=grpc,paths=source_relative:.  --go-grpc_out=require_unimplemented_servers=false,paths=source_relative:. ./types_wrapper.proto

/github.com/gogo/protobuf:/Users/joker/Desktop/gopath/src/github.com/gogo/protobuf/proto:.