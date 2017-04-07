package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"golang.org/x/net/context"

	"google.golang.org/grpc"

	"github.com/knq/envcfg"
	"github.com/knq/firebase"

	pb "github.com/rnd/kudu-proto/item"
	pt "github.com/rnd/kudu-proto/types"
)

var (
	port = flag.Int("port", 50051, "Item server port")
)

// Item represents firebase database model for item database ref.
type Item struct {
	Goal    string                   `json:"goal"`
	Tag     string                   `json:"tag"`
	Notes   string                   `json:"notes"`
	NotesMD string                   `json:"notes_md"`
	Created firebase.ServerTimestamp `json:"created"`
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

	items := make(map[string]Item)
	err = s.itemRef.Ref("/item").Get(&items)
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
func (s *server) AddItem(ctx context.Context, req *pb.AddRequest) (*pb.AddResponse, error) {
	var err error
	var res pb.AddResponse

	item := &Item{
		Goal:  req.Item.Goal,
		Tag:   req.Item.Tag,
		Notes: req.Item.Notes,
	}
	id, err := s.itemRef.Ref("/item").Push(item)
	if err != nil {
		log.Fatal(err)
	}
	res.Id = id
	return &res, nil
}

// GetItem get single item that matches with provided criteria.
func (s *server) GetItem(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	var err error
	var res pb.GetResponse

	var item Item
	err = s.itemRef.Ref("/item/" + req.Id).Get(&item)
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

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	server := grpc.NewServer()

	pb.RegisterItemServiceServer(server, newServer())
	server.Serve(lis)
}
