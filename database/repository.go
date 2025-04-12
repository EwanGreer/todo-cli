package database

import (
	"log"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Repository struct {
	*gorm.DB
}

func NewDatabase() (*Repository, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(homeDir, "task.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&List{}, &Task{}); err != nil {
		return nil, err
	}

	return &Repository{db}, nil
}

func (d *Repository) Save(task *Task) {
	tx := d.DB.Save(task)
	if tx.Error != nil {
		log.Println(tx.Error)
	}
}
