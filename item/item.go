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

// Item represents firebase database model for item data ref.
type Item struct {
	Goal      string `json:"goal"`
	URL       string `json:"url,omitempty"`
	Tags      string `json:"tags,omitempty"`
	Notes     string `json:"notes,omitempty"`
	NotesMD   string `json:"notes_md,omitempty"`
	Date      string `json:"date"`
	Completed bool   `json:"completed"`
}

// service is implementation of item service server.
type service struct {
	// config is server environment config.
	config *envcfg.Envcfg

	// dataRef is kudu-data firebase database ref.
	dataRef *firebase.DatabaseRef
}

// TodayItems retrieves all items added today.
func (s *service) TodayItems(ctx context.Context, req *pb.TodayItemsRequest) (*pb.TodayItemsResponse, error) {
	var err error
	res := &pb.TodayItemsResponse{
		Status: pb.ResponseStatus_INTERNAL_ERROR,
	}

	userID := ctx.Value(auth.UserIDKey).(string)
	today := time.Now().Format("20060102")
	items := make(map[string]Item)

	err = s.dataRef.Ref("/item/"+userID).Get(&items,
		firebase.OrderBy("date"),
		firebase.StartAt(today),
		firebase.EndAt(today),
	)

	log.Printf("filtering today items: %s", today)
	if err != nil {
		return res, err
	}

	for _, item := range items {
		res.Items = append(res.Items, &pb.Item{
			Goal:    item.Goal,
			Url:     item.URL,
			Tags:    item.Tags,
			Notes:   item.Notes,
			NotesMd: item.NotesMD,
		})
	}
	res.Status = pb.ResponseStatus_SUCCESS
	return res, nil
}

// AddItem add new item to datebase.
func (s *service) AddItem(ctx context.Context, req *pb.AddRequest) (*pb.AddResponse, error) {
	var err error
	res := &pb.AddResponse{
		Status: pb.ResponseStatus_INTERNAL_ERROR,
	}

	userID := ctx.Value(auth.UserIDKey).(string)
	today := time.Now().Format("20060102")
	item := &Item{
		Date:  today,
		Goal:  req.Item.Goal,
		URL:   req.Item.Url,
		Tags:  req.Item.Tags,
		Notes: req.Item.Notes,
	}

	id, err := s.dataRef.Ref("/item/" + userID).Push(item)
	if err != nil {
		return res, err
	}

	return &pb.AddResponse{
		Id:     id,
		Status: pb.ResponseStatus_SUCCESS,
	}, nil
}

// RemoveItem get single item that matches with provided criteria.
func (s *service) RemoveItem(ctx context.Context, req *pb.RemoveRequest) (*pb.RemoveResponse, error) {
	var err error
	res := &pb.RemoveResponse{
		Status: pb.ResponseStatus_INTERNAL_ERROR,
	}

	userID := ctx.Value(auth.UserIDKey).(string)
	path := fmt.Sprintf("/item/%s/%s", userID, req.Id)

	err = s.dataRef.Ref(path).Remove()
	if err != nil {
		return res, err
	}

	return &pb.RemoveResponse{
		Status: pb.ResponseStatus_SUCCESS,
	}, nil
}
