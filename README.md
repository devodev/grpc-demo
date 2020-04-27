# gRPC Demo

## Build
### Protocol buffers
Generate demo protocol buffers and output into demo package
```
$ protoc --proto_path=internal/demo --go_out=plugins=grpc:internal/demo --go_opt=paths=source_relative internal/demo/*.proto
```
