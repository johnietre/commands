package main

// Task holds information on a task
type Task struct {
  Name string `json:"name"`
  Description string `json:"description"`
  CompletedAt int64 `json:"completedAt"`
}

func CreateTask(name, desc string) *Task {
  return &Task{
    Name: name,
    Description: desc,
  }
}


