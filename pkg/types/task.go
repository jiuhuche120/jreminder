package types

import "time"

type Task struct {
	ID      string `json:"_id"`
	Content string `json:"content"`
}

type SubTask struct {
	ID        string    `json:"_id"`
	Content   string    `json:"content"`
	Executor  Executor  `json:"executor"`
	IsDone    bool      `json:"isDone"`
	StartDate time.Time `json:"startDate"`
	DueDate   time.Time `json:"dueDate"`
	Url       string
	DingTalk  string
}

type Executor struct {
	ID   string `json:"_id"`
	Name string `json:"name"`
}
