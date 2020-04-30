# gRPC Demo

## Build
### Protocol buffers
Generate demo protocol buffers and output into demo package
```
$ protoc --proto_path=internal/pb --go_out=plugins=grpc:internal/pb --go_opt=paths=source_relative internal/pb/*.proto
```

## Roadmap
### Security
Secure Layer
- Add TLS for all communication (token/websocket endpoints)
Authentication
- Add token endpoint (either different server or on the hub)
  - Let clients authenticate themselves using credentials or a certificate
  - On successful authentication, receive JWT token
  - Use JWT token as query param when dialing websocket endpoint
    - ws://localhost:8080/ws?token=UG2345nKJB...==
