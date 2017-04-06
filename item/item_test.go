package item

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/knq/firebase"
	pb "github.com/rnd/kudu-proto/item"
	pt "github.com/rnd/kudu-proto/types"
)

var server *Server
var defaultContext, defaultCancel = context.WithCancel(context.Background())

func mockData() {
	var req pb.AddRequest

	testData := []pt.Item{
		{
			Goal:  "Foo",
			Tag:   "Bar",
			Notes: "# Baz",
		},
		{
			Goal:  "Kudu",
			Tag:   "App",
			Notes: "## Test",
		},
	}

	for _, test := range testData {
		req.Item = &test
		_, err := server.AddItem(defaultContext, &req)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func clearData() {
	var err error

	keys := make(map[string]interface{})
	err = server.itemRef.Ref("/item").Get(&keys, firebase.Shallow)
	if err != nil {
		log.Fatal(err)
	}

	for key := range keys {
		err = server.itemRef.Ref("/item/" + key).Remove()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func TestMain(m *testing.M) {
	var err error

	server, err = New()
	if err != nil {
		log.Fatal(err)
	}

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
	}
	res, err := server.ListItem(defaultContext, &test)
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Items) != 2 {
		t.Errorf("Expected list to have 2 items, got: %d", len(res.Items))
	}
}

func TestAddItem(t *testing.T) {
	var err error
	var req pb.AddRequest

	req.Item = &pt.Item{
		Goal:  "Foo",
		Tag:   "Bar",
		Notes: "# Baz",
	}
	res, err := server.AddItem(defaultContext, &req)
	if err != nil {
		t.Fatal(err)
	}

	if res.Id == "" {
		t.Error("Expected id to not empty")
	}
}

func TestGetItem(t *testing.T) {
	var err error

	keys := make(map[string]interface{})
	err = server.itemRef.Ref("/item").Get(&keys, firebase.Shallow)
	if err != nil {
		t.Fatal("Failed to get item keys")
	}

	for key := range keys {
		res, err := server.GetItem(defaultContext, &pb.GetRequest{Id: key})
		if err != nil {
			t.Errorf("Got error on get item with key: %s, %v", key, err)
		}
		if res.Item == nil {
			t.Error("Expected response item to not empty")
		}
	}
}
