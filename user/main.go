package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/knq/envcfg"
	"github.com/knq/firebase"
	pb "github.com/rnd/kudu/golang/protogen/user"
)

// newService creates new instance of user service server.
func newService() *service {
	s := new(service)

	// setup config
	config, err := envcfg.New()
	if err != nil {
		log.Fatal(err)
	}
	s.config = config

	// setup database
	s.authRef, err = firebase.NewDatabaseRef(
		firebase.GoogleServiceAccountCredentialsJSON([]byte(s.config.GetKey("firebase.authcreds"))),
	)
	if err != nil {
		log.Fatal(err)
	}
	s.dataRef, err = firebase.NewDatabaseRef(
		firebase.GoogleServiceAccountCredentialsJSON([]byte(s.config.GetKey("firebase.datacreds"))),
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

	server := grpc.NewServer()
	pb.RegisterUserServiceServer(server, service)
	server.Serve(lis)
}
