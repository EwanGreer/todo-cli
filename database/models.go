package database

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Task struct {
	gorm.Model
	Name        string
	Description string // TODO: validate this is always at least X chars long
	Done        bool
}

type Database struct {
	*gorm.DB
}

func NewDatabase() (*Database, error) {
	db, err := gorm.Open(sqlite.Open("task.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&Task{}); err != nil {
		return nil, err
	}

	return &Database{db}, nil
}

func (d *Database) Save(task *Task) {
	tx := d.DB.Save(task)
	if tx.Error != nil {
		log.Println(tx.Error) // TODO: use slog global logger out to a file
	}
}
