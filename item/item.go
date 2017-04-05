package item

import (
	"context"
	"log"

	"github.com/knq/envcfg"
	"github.com/knq/firebase"

	pb "github.com/rnd/kudu-proto/item"
	pt "github.com/rnd/kudu-proto/types"
)

// Item represents firebase database model for item database ref.
type Item struct {
	Goal    string                   `json:"goal"`
	Tag     string                   `json:"tag"`
	Notes   string                   `json:"notes"`
	NotesMD string                   `json:"notes_md"`
	Created firebase.ServerTimestamp `json:"created"`
}

// Server is gRPC server.
type Server struct {
	// config is server environment config.
	config *envcfg.Envcfg

	// itemRef is firebase item database ref.
	itemRef *firebase.DatabaseRef
}

// New creates new instance of item server.
func New() (*Server, error) {
	var err error

	config, err := envcfg.New()
	if err != nil {
		return nil, err
	}

	itemRef, err := firebase.NewDatabaseRef(
		firebase.GoogleServiceAccountCredentialsJSON([]byte(config.GetKey("firebase.item.cred"))),
	)
	if err != nil {
		return nil, err
	}

	return &Server{
		config:  config,
		itemRef: itemRef,
	}, nil
}

// ListItem get list of item that matches with provided criteria.
func (s *Server) ListItem(ctx context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	var err error
	var res pb.ListResponse

	items := make(map[string]Item)
	err = s.itemRef.Ref("/item").Get(items)
	if err != nil {
		log.Fatal(err)
	}

	for _, item := range items {
		res.Items = append(res.Items, &pt.Item{
			Goal:    item.Goal,
			Tag:     item.Tag,
			Notes:   item.Notes,
			NotesMd: item.NotesMD,
		})
	}
	return &res, nil
}

// AddItem add new item to datebase.
func (s *Server) AddItem(ctx context.Context, req *pb.AddRequest) (*pb.AddResponse, error) {
	var err error
	var res pb.AddResponse

	item := &Item{
		Goal:  req.Item.Goal,
		Tag:   req.Item.Tag,
		Notes: req.Item.Notes,
	}
	_, err = s.itemRef.Ref("/item").Push(item)
	if err != nil {
		log.Fatal(err)
	}
	return &res, nil
}

// GetItem get single item that matches with provided criteria.
func (s *Server) GetItem(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	var err error
	var res pb.GetResponse

	var item Item
	err = s.itemRef.Ref("/item" + req.Id).Get(&item)
	if err != nil {
		log.Fatal(err)
	}

	res.Item = &pt.Item{
		Goal:    item.Goal,
		Tag:     item.Tag,
		Notes:   item.Notes,
		NotesMd: item.NotesMD,
	}
	return &res, nil
}
