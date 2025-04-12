package database

import (
	"github.com/EwanGreer/todo-cli/internal/status"
	"gorm.io/gorm"
)

type List struct {
	gorm.Model
	Name  string `gorm:"uniqueIndex"`
	Tasks []Task
}

type Task struct {
	gorm.Model
	Name        string        `gorm:"index"`
	Description string        // TODO: validate this is always at least X chars long
	Status      status.Status `gorm:"index"`
	ListID      uint
}

func (t Task) String() string {
	return t.Name
}

func (l List) String() string {
	return l.Name
}

func NewList(name string) *List {
	return &List{
		Name:  name,
		Tasks: []Task{},
	}
}

func NewTask(name string, desc string, status status.Status, parentID uint) *Task {
	return &Task{
		Name:        name,
		Description: desc,
		Status:      status,
		ListID:      parentID,
	}
}
