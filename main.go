package main

import (
	"log"
	"os"

	"github.com/EwanGreer/todo-cli/database"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	if os.Getenv("ENV") == "development" {
		f, err := tea.LogToFile("logs.log", "debug |")
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
	}

	db, err := database.NewDatabase()
	if err != nil {
		log.Fatal(err)
	}

	p := tea.NewProgram(initialModel(db))
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
