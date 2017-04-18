package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/rnd/kudu-service/auth"
	pb "github.com/rnd/kudu/golang/protogen/user"
)

func main() {
	service := newService()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", service.config.GetInt("server.port")))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var opts []grpc.ServerOption
	opts = append(opts, grpc.UnaryInterceptor(auth.UnaryInterceptor()))
	server := grpc.NewServer(opts...)

	pb.RegisterUserServiceServer(server, service)
	server.Serve(lis)
}
