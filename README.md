# gRPC Demo

## Build
### Protocol buffers
Generate demo protocol buffers and output into demo package
```
$ protoc --proto_path=internal/pb --go_out=plugins=grpc:internal/pb --go_opt=paths=source_relative internal/pb/*.proto
```
