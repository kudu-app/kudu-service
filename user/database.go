package main

import (
	"log"

	"github.com/knq/firebase"
)

// setupDatabase initialize firebase database ref.
func setupDatabase(s *service) {
	var err error

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
}
