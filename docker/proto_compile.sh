docker compose exec edge protoc -I=/proto --go_out=/src/proto --go-grpc_out=/src/proto /proto/*/*
docker compose exec load_balancer protoc -I=/proto --go_out=/src/proto --go-grpc_out=/src/proto /proto/LoadBalancer.proto
docker compose exec client python -m grpc_tools.protoc -I/proto --python_out=/src/proto --pyi_out=/src/proto --grpc_python_out=/src/proto /proto/*/*