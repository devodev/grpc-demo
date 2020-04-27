package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strconv"

	pb "github.com/devodev/grpc-demo/internal/pb"
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
	case "fluentd":
		if flag.NArg() < 3 {
			log.Fatal("fluentd command requires a method and timeout argument")
		}
		timeoutSec, err := strconv.Atoi(flag.Arg(2))
		if err != nil {
			log.Fatal(err)
		}

		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			log.Fatal(err)
		}
		client := pb.NewFluentdClient(conn)

		switch flag.Arg(1) {
		default:
			log.Fatal("unsupported method")
		case "start":
			req := &pb.FluentdStartRequest{TimeoutSec: int32(timeoutSec)}
			resp, err := client.Start(context.Background(), req)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Start response: %v\n", resp.GetStatus())
		case "stop":
			req := &pb.FluentdStopRequest{TimeoutSec: int32(timeoutSec)}
			resp, err := client.Stop(context.Background(), req)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Stop response: %v\n", resp.GetStatus())
		case "restart":
			req := &pb.FluentdRestartRequest{TimeoutSec: int32(timeoutSec)}
			resp, err := client.Restart(context.Background(), req)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Restart response: %v\n", resp.GetStatus())
		}

	}
}
