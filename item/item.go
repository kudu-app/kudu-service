package main

import (
	"fmt"
	"log"
	"time"

	"golang.org/x/net/context"

	"github.com/knq/envcfg"
	"github.com/knq/firebase"

	"github.com/rnd/kudu-service/auth"
	pb "github.com/rnd/kudu/golang/protogen/item"
)

const itemRef = "/data/%s/items/"

// Item represents firebase database model for item database ref.
type Item struct {
	Goal    string                   `json:"goal"`
	URL     string                   `json:"url"`
	Tag     string                   `json:"tag"`
	Notes   string                   `json:"notes"`
	NotesMD string                   `json:"notes_md"`
	Created firebase.ServerTimestamp `json:"created"`
}

// service is implementation of item service server.
type service struct {
	// config is server environment config.
	config *envcfg.Envcfg

	// itemRef is firebase item database ref.
	itemRef *firebase.DatabaseRef
}

// newService creates new instance of item service server.
func newService() *service {
	config, err := envcfg.New()
	if err != nil {
		log.Fatal(err)
	}
	itemRef, err := firebase.NewDatabaseRef(
		firebase.GoogleServiceAccountCredentialsJSON([]byte(config.GetKey("firebase.itemcreds"))),
	)
	if err != nil {
		log.Fatal(err)
	}
	return &service{
		config:  config,
		itemRef: itemRef,
	}
}

// ListItem get list of item that matches with provided criteria.
func (s *service) ListItem(ctx context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	var err error
	var res pb.ListResponse

	userId := ctx.Value(auth.UserIDKey).(string)
	path := fmt.Sprintf(itemRef, userId)
	date := time.Date(
		int(req.Date.GetYear()),
		time.Month(req.Date.GetMonth()),
		int(req.Date.GetDay()),
		0, 0, 0, 0,
		time.UTC,
	)

	items := make(map[string]Item)
	err = s.itemRef.Ref(path + date.Format("20060102")).Get(&items)
	if err != nil {
		log.Fatal(err)
	}

	for _, item := range items {
		res.Items = append(res.Items, &pb.Item{
			Goal:    item.Goal,
			Url:     item.URL,
			Tag:     item.Tag,
			Notes:   item.Notes,
			NotesMd: item.NotesMD,
		})
	}
	return &res, nil
}

// AddItem add new item to datebase.
func (s *service) AddItem(ctx context.Context, req *pb.AddRequest) (*pb.AddResponse, error) {
	var err error
	var res pb.AddResponse

	userId := ctx.Value(auth.UserIDKey).(string)
	path := fmt.Sprintf(itemRef, userId)
	date := time.Date(
		int(req.Item.Date.GetYear()),
		time.Month(req.Item.Date.GetMonth()),
		int(req.Item.Date.GetDay()),
		0, 0, 0, 0,
		time.UTC,
	)

	item := &Item{
		Goal:  req.Item.Goal,
		URL:   req.Item.Url,
		Tag:   req.Item.Tag,
		Notes: req.Item.Notes,
	}
	id, err := s.itemRef.Ref(path + date.Format("20060102")).Push(item)
	if err != nil {
		return nil, err
	}
	res.Id = id
	return &res, nil
}

// GetItem get single item that matches with provided criteria.
func (s *service) GetItem(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	var err error
	var res pb.GetResponse

	userId := ctx.Value(auth.UserIDKey).(string)
	path := fmt.Sprintf(itemRef, userId)

	var item Item
	err = s.itemRef.Ref(path + req.Id).Get(&item)
	if err != nil {
		log.Fatal(err)
	}

	res.Item = &pb.Item{
		Goal:    item.Goal,
		Url:     item.URL,
		Tag:     item.Tag,
		Notes:   item.Notes,
		NotesMd: item.NotesMD,
	}
	return &res, nil
}
