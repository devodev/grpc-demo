# gRPC Demo

## Build
### Protocol buffers
```
$ protoc --proto_path=demo --go_out=build/gen --go_opt=paths=source_relative demo/demo.proto
```
