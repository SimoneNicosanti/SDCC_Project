go get
protoc -I=/proto --go_out=/src/proto --go-grpc_out=/src/proto /proto/Login.proto
go run main.go