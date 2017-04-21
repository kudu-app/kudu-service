package main

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	"github.com/knq/envcfg"
	"github.com/knq/firebase"

	"github.com/rnd/kudu-service/auth"
	pb "github.com/rnd/kudu/golang/protogen/task"
)

// Task represents firebase database model for task data ref.
type Task struct {
	TaskName  string `json:"task_name"`
	URL       string `json:"url,omitempty"`
	Tags      string `json:"tags,omitempty"`
	Notes     string `json:"notes,omitempty"`
	NotesMD   string `json:"notes_md,omitempty"`
	Date      string `json:"date"`
	Completed bool   `json:"completed"`
}

// service is implementation of task service server.
type service struct {
	// config is server environment config.
	config *envcfg.Envcfg

	// dataRef is kudu-data firebase database ref.
	dataRef *firebase.DatabaseRef
}

// TodayTasks retrieves all tasks added today.
func (s *service) TodayTasks(ctx context.Context, req *pb.TodayTasksRequest) (*pb.TodayTasksResponse, error) {
	var err error
	res := &pb.TodayTasksResponse{
		Status: pb.ResponseStatus_INTERNAL_ERROR,
	}

	userID := ctx.Value(auth.UserIDKey).(string)
	today := time.Now().Format("20060102")
	tasks := make(map[string]Task)

	err = s.dataRef.Ref("/task/"+userID).Get(&tasks,
		firebase.OrderBy("date"),
		firebase.StartAt(today),
		firebase.EndAt(today),
	)

	if err != nil {
		return res, err
	}

	for _, task := range tasks {
		res.Tasks = append(res.Tasks, &pb.Task{
			TaskName: task.TaskName,
			Url:      task.URL,
			Tags:     task.Tags,
			Notes:    task.Notes,
			NotesMd:  task.NotesMD,
		})
	}
	res.Status = pb.ResponseStatus_SUCCESS
	return res, nil
}

// AddTask add new task to datebase.
func (s *service) AddTask(ctx context.Context, req *pb.AddRequest) (*pb.AddResponse, error) {
	var err error
	res := &pb.AddResponse{
		Status: pb.ResponseStatus_INTERNAL_ERROR,
	}

	userID := ctx.Value(auth.UserIDKey).(string)
	today := time.Now().Format("20060102")
	task := &Task{
		Date:     today,
		TaskName: req.Task.TaskName,
		URL:      req.Task.Url,
		Tags:     req.Task.Tags,
		Notes:    req.Task.Notes,
	}

	id, err := s.dataRef.Ref("/task/" + userID).Push(task)
	if err != nil {
		return res, err
	}

	return &pb.AddResponse{
		Id:     id,
		Status: pb.ResponseStatus_SUCCESS,
	}, nil
}

// RemoveTask get single task that matches with provided criteria.
func (s *service) RemoveTask(ctx context.Context, req *pb.RemoveRequest) (*pb.RemoveResponse, error) {
	var err error
	res := &pb.RemoveResponse{
		Status: pb.ResponseStatus_INTERNAL_ERROR,
	}

	userID := ctx.Value(auth.UserIDKey).(string)
	path := fmt.Sprintf("/task/%s/%s", userID, req.Id)

	err = s.dataRef.Ref(path).Remove()
	if err != nil {
		return res, err
	}

	return &pb.RemoveResponse{
		Status: pb.ResponseStatus_SUCCESS,
	}, nil
}
