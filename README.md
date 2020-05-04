# gRPC Demo

## Context
- We own SOCs that hosts our application environment.
- We own servers residing in client remote private networks that are behind corporate firewalls.
  - We cannot reach remote servers first because of NAT/Firewalls. The server needs to initiate the connection first.
- We need to communicate with these remote servers from the SOCs.

## Initial Idea
- Remote servers are plain gRPC server.
  - Easy to work with.
  - Easy to update independently.
  - No dependencies other than gRPC itself.
- Hub is a proxy between the SOC and remote servers.
  - Accept incoming connection from remote servers and register them.
- Client is a CLI that communicates with the Hub using plain gRPC.

## Concept
- The remote server creates a gRPC server.
- The remote server then initiates a Websocket connection to a "Hub" server.
- The remote server does:
  - It wraps the Websocket connection to convert it to a raw socket.
  - It serves the gRPC server using the wrapped connection as listener.
- The Hub does:
  - Authentication/authorization, etc.
  - Wraps the websocket connection to convert it to a raw socket.
  - Assigns the connection a unique ID and registers it internally.
- The Hub is now free to use the registered connection as a dialer when making gRPC requests.

## Issues
- I have not been successful trying to reroute gRPC incoming calls to registered connections other than:
  - Accepting a raw connection on a TCP port and directly connecting both ends with a Pipe to the registered connection.

## Next steps
- Try to implement a gRPC server on the hub and apply custom handlers returned by a director implementation using grpc-proxy package.
  - This would let us control incoming calls and create our own handshake protocl using gRPC metadata.

## Implementation
![Implementation Diagram 1](assets/img/implementation_diagram_1.png "Implementation Diagram 1")
![Implementation Sequence Diagram 1](assets/img/implementation_sequence_diagram_1.png "Implementation Sequence Diagram 1")

### Server Flow
- Dial the hub on its websocket listening uri (wss://hub:8080/ws).
- Upon successful websocket upgrade, it creates a gRPC server, wraps the websocket connection into a raw socket and listens on it.

### Hub Flow
Remote Server side
- Listen on wss://hub:8080/ws for incoming websocket connections.
- Upon successful websocket upgrade, it wraps the connection and registers it using a unique ID.

Local client side
- Listen on port 9090 for raw TCP connections.
  - The client can dial this raw socket and use the connection returned as a dialer for its gRPC call. This is the current implementation.
- Upon successful TCP accept, it connects both ends to a registered connection.
  - Currently, for demonstration purposes, the first registered connection is hardcoded as the one served.
  - My idea would be to do a custom handshake on TCP using a JSON payload between the client and hub.
    - This handshake would let the client send authentication details, and also asks for a particular server to talk to.

### Client Flow
- Provides a CLI to make gRPC requests to a gRPC server.
- Can provide a request payload using STDIN or a file.
- Codecs available: json/xml/yaml.


## Development
Generate cert/key in cwd for hub (test.crt/test.key)
```
$ make gencert
$ make readcert
reading generated cert..
openssl x509 -text -noout -in test.crt
Certificate:
...
```

Regenerate protocol buffers
```
$ make pb
```

Build all
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
