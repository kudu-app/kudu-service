package main

import (
	"log"
	"os"
	"testing"

	"golang.org/x/net/context"

	"github.com/knq/firebase"
	"github.com/rnd/kudu-service/auth"
	pb "github.com/rnd/kudu/golang/protogen/item"
)

var testService *service
var defaultContext context.Context
var defaultCancel context.CancelFunc
var userID string

func mockData() {
	var req pb.AddRequest

	testData := []pb.Item{
		{
			Goal:  "Foo",
			Tags:  "Bar",
			Notes: "# Baz",
			Url:   "brank.as",
		},
		{
			Goal:  "Kudu",
			Tags:  "App",
			Notes: "## Test",
			Url:   "google.com",
		},
	}

	for _, test := range testData {
		req.Item = &test
		_, err := testService.AddItem(defaultContext, &req)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func clearData() {
	var err error

	err = testService.dataRef.Ref("/item").Remove()
	if err != nil {
		log.Fatal(err)
	}
}

func TestMain(m *testing.M) {
	testService = newService()

	//TODO: Replace dummy user ID.
	userID = "foo"
	defaultContext, defaultCancel = context.WithCancel(context.WithValue(context.Background(), auth.UserIDKey, userID))

	clearData()
	mockData()

	code := m.Run()

	os.Exit(code)
}

func TestTodayItems(t *testing.T) {
	var err error

	test := pb.TodayItemsRequest{
		Goal: "Foo",
		Tags: "Bar",
	}
	res, err := testService.TodayItems(defaultContext, &test)
	if err != nil {
		t.Fatal(err)
	}

	if res.Status != pb.ResponseStatus_SUCCESS {
		t.Fatalf("expected response status is: '%v', got: '%v'",
			pb.ResponseStatus_SUCCESS,
			res.Status)
	}
}

func TestAddItem(t *testing.T) {
	var err error
	var req pb.AddRequest

	req.Item = &pb.Item{
		Goal:  "Foo",
		Tags:  "Bar",
		Notes: "# Baz",
		Url:   "reddit.com",
	}
	res, err := testService.AddItem(defaultContext, &req)
	if err != nil {
		t.Fatal(err)
	}

	if res.Status != pb.ResponseStatus_SUCCESS {
		t.Fatalf("expected response status is: '%v', got: '%v'",
			pb.ResponseStatus_SUCCESS,
			res.Status)
	}

	if res.Id == "" {
		t.Error("Expected id to not empty")
	}
}

func TestRemoveItem(t *testing.T) {
	var err error

	keys := make(map[string]interface{})
	err = testService.dataRef.Ref("/item/"+userID).Get(&keys, firebase.Shallow)
	if err != nil {
		log.Fatal(err)
	}

	if len(keys) < 1 {
		log.Fatalf("expected at least one item to be present")
	}

	for key := range keys {
		res, err := testService.RemoveItem(defaultContext, &pb.RemoveRequest{Id: key})
		if err != nil {
			t.Fatal(err)
		}

		if res.Status != pb.ResponseStatus_SUCCESS {
			t.Fatalf("expected response status is: '%v', got: '%v'",
				pb.ResponseStatus_SUCCESS,
				res.Status)
		}
	}
}
