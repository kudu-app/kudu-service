package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"

	"github.com/knq/envcfg"
	"github.com/knq/firebase"

	"github.com/rnd/kudu-service/auth/token"
	pb "github.com/rnd/kudu/golang/protogen/user"
)

var (
	// ErrInvalidCredential is the error returned when the credentials
	// is not match.
	ErrInvalidCredential = errors.New("Invalid username or password")
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

// Register set new user record into user database ref.
func (s *service) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	var err error
	res := &pb.RegisterResponse{
		Status: pb.ResponseStatus_INTERNAL_ERROR,
	}

	encEmail := base64.StdEncoding.EncodeToString([]byte(req.Credential.Email))
	encUsername := base64.StdEncoding.EncodeToString([]byte(req.Credential.Username))

	// Create user auth record
	userID, err := s.authRef.Ref("/user").Push(map[string]interface{}{
		"credentials": map[string]interface{}{
			"email": map[string]interface{}{
				encEmail: true,
			},
			"username": map[string]interface{}{
				encUsername: true,
			},
		},
	})
	if err != nil {
		return res, err
	}

	// Hash password
	passBuf, err := bcrypt.GenerateFromPassword([]byte(req.Credential.Password), bcrypt.DefaultCost)
	if err != nil {
		return res, err
	}

	// Set email credential
	err = s.authRef.Ref("/credential/email").Update(map[string]interface{}{
		encEmail: map[string]interface{}{
			"secret":  string(passBuf),
			"user_id": userID,
		},
	})
	if err != nil {
		return res, err
	}

	// Set username credential
	err = s.authRef.Ref("/credential/username").Update(map[string]interface{}{
		encUsername: map[string]interface{}{
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
	encEmail := base64.StdEncoding.EncodeToString([]byte(req.Credential.Email))
	var creds struct {
		UserID string `json:"user_id"`
		Secret string `json:"secret"`
	}
	err = s.authRef.Ref("/credential/email/" + encEmail).Get(&creds)
	if err != nil {
		return res, err
	}

	if creds.UserID == "" {
		res.Status = pb.ResponseStatus_CREDENTIAL_INVALID
		return res, ErrInvalidCredential
	}

	// validate credential
	err = bcrypt.CompareHashAndPassword([]byte(creds.Secret), []byte(req.Credential.Password))
	if err != nil {
		res.Status = pb.ResponseStatus_CREDENTIAL_INVALID
		return res, ErrInvalidCredential
	}

	var user User
	err = s.dataRef.Ref("/user/" + creds.UserID).Get(&user)
	if err != nil {
		return res, err
	}

	// generate token
	exp := time.Now().Add(token.DefaultExp)
	claims := &token.Claims{
		User: token.User{
			ID:          creds.UserID,
			DisplayName: fmt.Sprintf("%s %s", user.FirstName, user.LastName),
		},
		IssuedAt:   json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
		Expiration: json.Number(strconv.FormatInt(exp.Unix(), 10)),
	}

	token, err := token.New(claims,
		s.config.GetKey("jwt.privatekey"),
		s.config.GetKey("jwt.publickey"),
	)
	if err != nil {
		return res, err
	}

	return &pb.LoginResponse{
		Status: pb.ResponseStatus_SUCCESS,
		Token:  token,
	}, nil
}
