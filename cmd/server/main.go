package main

import (
	"context"
	"flag"
	"log"
	"net"

	"github.com/devodev/grpc-demo/internal/demo"
	"google.golang.org/grpc"
)

type helloServiceServer struct {
}

func (s *helloServiceServer) GetHello(ctx context.Context, payload *demo.HelloRequest) (*demo.HelloResponse, error) {
	message := "Hello " + payload.GetName()
	return &demo.HelloResponse{Message: message}, nil
}

func main() {
	var (
		addr string
	)
	flag.StringVar(&addr, "listen", ":9300", "listening address.")
	flag.Parse()

	server := grpc.NewServer()
	helloService := &helloServiceServer{}
	demo.RegisterHelloServiceServer(server, helloService)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(server.Serve(l))
}
