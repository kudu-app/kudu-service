package main

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/knq/firebase"
	"github.com/rnd/kudu-service/auth"
	pb "github.com/rnd/kudu/golang/protogen/item"
	pdate "github.com/rnd/kudu/golang/protogen/type/date"
)

var testServer *server
var defaultContext context.Context
var defaultCancel context.CancelFunc
var userID string
var now = time.Now()

func mockData() {
	var req pb.AddRequest

	testData := []pb.Item{
		{
			Goal:  "Foo",
			Tag:   "Bar",
			Notes: "# Baz",
			Url:   "brank.as",
			Date: &pdate.Date{
				Year:  int32(now.Year()),
				Month: int32(now.Month()),
				Day:   int32(now.Day()),
			},
		},
		{
			Goal:  "Kudu",
			Tag:   "App",
			Notes: "## Test",
			Url:   "google.com",
			Date: &pdate.Date{
				Year:  int32(now.Year()),
				Month: int32(now.Month()),
				Day:   int32(now.Day() + 1),
			},
		},
	}

	for _, test := range testData {
		req.Item = &test
		_, err := testServer.AddItem(defaultContext, &req)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func clearData() {
	var err error

	path := fmt.Sprintf(itemRef, userID)

	keys := make(map[string]interface{})
	err = testServer.itemRef.Ref(path).Get(&keys, firebase.Shallow)
	if err != nil {
		log.Fatal(err)
	}

	for key := range keys {
		err = testServer.itemRef.Ref(path + key).Remove()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func TestMain(m *testing.M) {
	testServer = newServer()

	//TODO: Fix this by fetch test user id.
	userID = "foo"
	defaultContext, defaultCancel = context.WithCancel(context.WithValue(context.Background(), auth.UserIDKey, userID))

	clearData()
	mockData()

	code := m.Run()

	os.Exit(code)
}

func TestListItem(t *testing.T) {
	var err error

	test := pb.ListRequest{
		Goal: "Foo",
		Tag:  "Bar",
		Date: &pdate.Date{
			Year:  int32(now.Year()),
			Month: int32(now.Month()),
			Day:   int32(now.Day()),
		},
	}
	res, err := testServer.ListItem(defaultContext, &test)
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Items) != 1 {
		t.Errorf("Expected list to have 1 items, got: %d", len(res.Items))
	}
}

func TestAddItem(t *testing.T) {
	var err error
	var req pb.AddRequest

	now := time.Now()
	req.Item = &pb.Item{
		Goal:  "Foo",
		Tag:   "Bar",
		Notes: "# Baz",
		Url:   "reddit.com",
		Date: &pdate.Date{
			Year:  int32(now.Year()),
			Month: int32(now.Month()),
			Day:   int32(now.Day() + 2),
		},
	}
	res, err := testServer.AddItem(defaultContext, &req)
	if err != nil {
		t.Fatal(err)
	}

	if res.Id == "" {
		t.Error("Expected id to not empty")
	}
}

func TestGetItem(t *testing.T) {
	var err error

	date := now.Format("20060102")
	res, err := testServer.GetItem(defaultContext, &pb.GetRequest{Id: date})
	if err != nil {
		t.Errorf("Got error on get item on date: %s, %v", date, err)
	}
	if res.Item == nil {
		t.Error("Expected response item to not empty")
	}
}
