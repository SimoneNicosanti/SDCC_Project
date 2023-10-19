go get
protoc -I=/proto --go_out=/src/proto --go-grpc_out=/src/proto /proto/LoadBalancer.proto
go build -o balancer.out main.go