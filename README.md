# gRPC Demo

## Development
Generate protocol buffers

Generate cert/key in cwd for hub (test.crt/test.key)
```
$ make
```

Run hub
```
$ go run ./cmd/hub serve --tls --tls-cert-file test.crt --tls-key-file test.key
```

Connect server to hub
```
$ go run ./cmd/server serve --tls-insecure-skip-verify --hub-uri wss://localhost:8080/ws
```

Call grpc service on the server through hub using raw tcp forward
```
$ echo '{}' | go run ./cmd/client fluentd start -s localhost:9090
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
