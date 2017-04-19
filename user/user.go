package main

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"

	"github.com/knq/envcfg"
	"github.com/knq/firebase"
	"github.com/knq/jwt"

	pb "github.com/rnd/kudu/golang/protogen/user"
)

var (
	InvalidCredential = errors.New("Invalid username or password")
)

// User represents firebase database model for user database ref.
type User struct {
	FirstName string                   `json:"first_name"`
	LastName  string                   `json:"last_name"`
	Created   firebase.ServerTimestamp `json:"created"`
}

// service is implementation of user service server.
type service struct {
	// config is server environment config.
	config *envcfg.Envcfg

	// authRef is kudu-auth firebase database ref.
	authRef *firebase.DatabaseRef

	// dataRef is kudu-data firebase database ref.
	dataRef *firebase.DatabaseRef
}

// newService creates new instance of user service server.
func newService() *service {
	s := new(service)

	config, err := envcfg.New()
	if err != nil {
		log.Fatal(err)
	}
	s.config = config

	setupDatabase(s)
	return s
}

// Register set new user record into user database ref.
func (s *service) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	var err error
	res := &pb.RegisterResponse{
		Status: pb.ResponseStatus_INTERNAL_ERROR,
	}

	// Create user auth record
	userID, err := s.authRef.Ref("/user").Push(map[string]interface{}{
		"credentials": map[string]interface{}{
			"email": map[string]interface{}{
				req.Credential.Email: true,
			},
			"username": map[string]interface{}{
				req.Credential.Username: true,
			},
		},
	})

	// Hash password
	passBuf, err := bcrypt.GenerateFromPassword([]byte(req.Credential.Password), bcrypt.DefaultCost)
	if err != nil {
		return res, err
	}

	// Set email credential
	err = s.authRef.Ref("/credential/email").Set(map[string]interface{}{
		req.Credential.Email: map[string]interface{}{
			"secret":  string(passBuf),
			"user_id": userID,
		},
	})
	if err != nil {
		return res, err
	}

	// Set username credential
	err = s.authRef.Ref("/credential/username").Set(map[string]interface{}{
		req.Credential.Username: map[string]interface{}{
			"secret":  string(passBuf),
			"user_id": userID,
		},
	})
	if err != nil {
		return res, err
	}

	// Create user data record
	user := &User{
		FirstName: req.User.FirstName,
		LastName:  req.User.LastName,
	}
	err = s.dataRef.Ref("/user/" + userID).Set(user)
	if err != nil {
		return res, err
	}

	return &pb.RegisterResponse{
		Status: pb.ResponseStatus_SUCCESS,
	}, nil
}

// Login validates user claims and generate jwt token.
func (s *service) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	var err error
	res := &pb.LoginResponse{
		Status: pb.ResponseStatus_INTERNAL_ERROR,
	}

	// TODO: Handle login with username

	// find email / username
	var creds struct {
		UserID string `json:"user_id"`
		Secret string `json:"secret"`
	}
	err = s.authRef.Ref("/credential/email/" + req.Credential.Email).Get(&creds)
	if err != nil {
		return res, err
	}

	if creds.UserID == "" {
		res.Status = pb.ResponseStatus_CREDENTIAL_INVALID
		return res, InvalidCredential
	}

	// validate credential
	err = bcrypt.CompareHashAndPassword([]byte(creds.Secret), []byte(req.Credential.Password))
	if err != nil {
		res.Status = pb.ResponseStatus_CREDENTIAL_INVALID
		return res, InvalidCredential
	}

	// generate jwt token
	es384, err := jwt.ES384.New(jwt.PEM{
		[]byte(s.config.GetKey("jwt.privatekey")),
		[]byte(s.config.GetKey("jwt.publickey")),
	})
	if err != nil {
		return res, err
	}

	expr := time.Now().Add(time.Minute * 15)
	claim := &jwt.Claims{
		Issuer:     creds.UserID,
		Expiration: json.Number(strconv.FormatInt(expr.Unix(), 10)),
	}

	tokenBuf, err := es384.Encode(claim)
	if err != nil {
		return res, err
	}

	return &pb.LoginResponse{
		Status: pb.ResponseStatus_SUCCESS,
		Token:  string(tokenBuf),
	}, nil
}
