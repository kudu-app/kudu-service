package main

import (
	"flag"
	"log"

	"golang.org/x/net/context"

	"github.com/knq/envcfg"
	"github.com/knq/firebase"

	pb "github.com/rnd/kudu/golang/protogen/user"
)

var (
	port = flag.Int("port", 50052, "User server port")
)

const userRef = "/user/%s"

// User represents firebase database model for user database ref.
type User struct {
	Username  string                   `json:"username"`
	FirstName string                   `json:"first_name"`
	LastName  string                   `json:"last_name"`
	Created   firebase.ServerTimestamp `json:"created"`
}

// server is gRPC server.
type server struct {
	// config is server environment config.
	config *envcfg.Envcfg

	// userRef is firebase user database ref.
	userRef *firebase.DatabaseRef
}

// newServer creates new instance of user server.
func newServer() *server {
	config, err := envcfg.New()
	if err != nil {
		log.Fatal(err)
	}
	userRef, err := firebase.NewDatabaseRef(
		firebase.GoogleServiceAccountCredentialsJSON([]byte(config.GetKey("firebase.usercreds"))),
	)
	if err != nil {
		log.Fatal(err)
	}
	return &server{
		config:  config,
		userRef: userRef,
	}
}

// Register set new user record into user database ref.
func (s *server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	var err error
	var res pb.RegisterResponse

	// user := &User{
	// 	Username:  req.User.Username,
	// 	FirstName: req.User.FirstName,
	// 	LastName:  req.User.LastName,
	// }

	// creds := &auth.Credential{
	// 	Email:    req.Credential.Email,
	// 	Password: req.Credential.Password,
	// }

	// TODO:
	// - check if username already taken
	// - check if email already taken
	// - set user
	// - set credential

	return &res, err
}

// Login validates user claims and generate jwt token.
func (s *server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	var err error
	var res pb.LoginResponse

	return &res, err
}
