package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/knq/envcfg"
	"github.com/knq/firebase"
	"github.com/rnd/kudu-service/auth"
	pb "github.com/rnd/kudu/golang/protogen/item"
)

// newService creates new instance of item service server.
func newService() *service {
	s := new(service)

	// setup config
	config, err := envcfg.New()
	if err != nil {
		log.Fatal(err)
	}

	// setup database
	s.dataRef, err = firebase.NewDatabaseRef(
		firebase.GoogleServiceAccountCredentialsJSON([]byte(config.GetKey("firebase.datacreds"))),
	)
	if err != nil {
		log.Fatal(err)
	}

	return s
}

func main() {
	service := newService()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", service.config.GetInt("server.port")))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var opts []grpc.ServerOption
	opts = append(opts, grpc.UnaryInterceptor(auth.UnaryInterceptor()))
	server := grpc.NewServer(opts...)

	pb.RegisterItemServiceServer(server, service)
	server.Serve(lis)
}
