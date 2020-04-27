package main

import (
	"flag"
	"log"
	"net"

	"google.golang.org/grpc"
)

func main() {
	var (
		addr string
	)
	flag.StringVar(&addr, "listen", ":9300", "listening address.")
	flag.Parse()

	server := grpc.NewServer()
	fluentdService := FluentdService{}
	fluentdService.RegisterServer(server)

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(server.Serve(l))
}
