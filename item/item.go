package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"golang.org/x/net/context"

	"github.com/knq/envcfg"
	"github.com/knq/firebase"

	"github.com/rnd/kudu-service/auth"
	pb "github.com/rnd/kudu/golang/protogen/item"
	pdate "github.com/rnd/kudu/golang/protogen/type/date"
)

var (
	port = flag.Int("port", 50051, "Item server port")
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
	Date    firebase.Time            `json:"date"`
}

// server is gRPC server.
type server struct {
	// config is server environment config.
	config *envcfg.Envcfg

	// itemRef is firebase item database ref.
	itemRef *firebase.DatabaseRef
}

// newServer creates new instance of item server.
func newServer() *server {
	var err error

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
	return &server{
		config:  config,
		itemRef: itemRef,
	}
}

// ListItem get list of item that matches with provided criteria.
func (s *server) ListItem(ctx context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	var err error
	var res pb.ListResponse

	userId := ctx.Value(auth.UserIDKey).(string)
	path := fmt.Sprintf(itemRef, userId)

	items := make(map[string]Item)
	err = s.itemRef.Ref(path).Get(&items)
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
			Date: &pdate.Date{
				Year:  int32(item.Date.Time().Year()),
				Month: int32(item.Date.Time().Month()),
				Day:   int32(item.Date.Time().Day()),
			},
		})
	}
	return &res, nil
}

// AddItem add new item to datebase.
func (s *server) AddItem(ctx context.Context, req *pb.AddRequest) (*pb.AddResponse, error) {
	var err error
	var res pb.AddResponse

	userId := ctx.Value(auth.UserIDKey).(string)
	path := fmt.Sprintf(itemRef, userId)

	date, err := time.Parse("20060102",
		fmt.Sprintf("%d%02d%02d",
			req.Item.Date.GetYear(),
			req.Item.Date.GetMonth(),
			req.Item.Date.GetDay()))
	if err != nil {
		return nil, err
	}

	item := &Item{
		Goal:  req.Item.Goal,
		URL:   req.Item.Url,
		Tag:   req.Item.Tag,
		Notes: req.Item.Notes,
		Date:  firebase.Time(date),
	}
	id, err := s.itemRef.Ref(path).Push(item)
	if err != nil {
		return nil, err
	}
	res.Id = id
	return &res, nil
}

// GetItem get single item that matches with provided criteria.
func (s *server) GetItem(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
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
		Date: &pdate.Date{
			Year:  int32(item.Date.Time().Year()),
			Month: int32(item.Date.Time().Month()),
			Day:   int32(item.Date.Time().Day()),
		},
	}
	return &res, nil
}
