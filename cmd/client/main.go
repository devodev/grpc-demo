package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/devodev/grpc-demo/internal/demo"
	"google.golang.org/grpc"
)

func main() {
	var (
		addr string
	)
	flag.StringVar(&addr, "addr", "localhost:9300", "gRPC server address and port to connect to")
	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatal("missing command")
	}

	cmd := flag.Arg(0)
	switch cmd {
	default:
		log.Fatal("command not supported")
	case "hello":
		if flag.NArg() < 2 {
			log.Fatal("hello command requires a name argument")
		}
		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			log.Fatal(err)
		}
		client := demo.NewHelloServiceClient(conn)

		req := &demo.HelloRequest{Name: flag.Arg(1)}
		resp, err := client.GetHello(context.Background(), req)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("GetHello response: %v\n", resp.GetMessage())
	}
}
