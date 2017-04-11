package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	pb "github.com/rnd/kudu-proto/item"
	"github.com/rnd/kudu-service/auth"
)

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var opts []grpc.ServerOption
	opts = append(opts, grpc.UnaryInterceptor(auth.UnaryInterceptor()))
	server := grpc.NewServer(opts...)

	pb.RegisterItemServiceServer(server, newServer())
	server.Serve(lis)
}
