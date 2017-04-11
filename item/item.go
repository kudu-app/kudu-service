package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	"github.com/knq/envcfg"
	"github.com/knq/firebase"
	"github.com/knq/jwt"

	pb "github.com/rnd/kudu-proto/item"
	pt "github.com/rnd/kudu-proto/types"
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

	userId := ctx.Value("userid").(string)
	path := fmt.Sprintf(itemRef, userId)

	items := make(map[string]Item)
	err = s.itemRef.Ref(path).Get(&items)
	if err != nil {
		log.Fatal(err)
	}

	for _, item := range items {
		res.Items = append(res.Items, &pt.Item{
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
func (s *server) AddItem(ctx context.Context, req *pb.AddRequest) (*pb.AddResponse, error) {
	var err error
	var res pb.AddResponse

	userId := ctx.Value("userid").(string)
	path := fmt.Sprintf(itemRef, userId)

	item := &Item{
		Goal:  req.Item.Goal,
		URL:   req.Item.Url,
		Tag:   req.Item.Tag,
		Notes: req.Item.Notes,
	}
	id, err := s.itemRef.Ref(path).Push(item)
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

	userId := ctx.Value("userid").(string)
	path := fmt.Sprintf(itemRef, userId)

	var item Item
	err = s.itemRef.Ref(path + req.Id).Get(&item)
	if err != nil {
		log.Fatal(err)
	}

	res.Item = &pt.Item{
		Goal:    item.Goal,
		Url:     item.URL,
		Tag:     item.Tag,
		Notes:   item.Notes,
		NotesMd: item.NotesMD,
	}
	return &res, nil
}

// authUnaryInterceptor is grpc middleware that responsible to handle authentication,
// by parsing the token field from metadata.
func authUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var err error

	md, ok := metadata.FromContext(ctx)
	if !ok {
		return nil, grpc.Errorf(codes.DataLoss, "auth unary interceptor: failed to get metadata")
	}

	var userId string
	if token, ok := md["token"]; ok {
		userId, err = jwt.PeekPayloadField([]byte(token[0]), "uid")
		if err != nil {
			return nil, err
		}
		newCtx := context.WithValue(ctx, "userid", userId)
		return handler(newCtx, req)
	}
	return nil, grpc.Errorf(codes.Unauthenticated, "authentication required")
}

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var opts []grpc.ServerOption
	opts = append(opts, grpc.UnaryInterceptor(authUnaryInterceptor))
	server := grpc.NewServer(opts...)

	pb.RegisterItemServiceServer(server, newServer())
	server.Serve(lis)
}
