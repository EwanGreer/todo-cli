package main

import (
	"fmt"
	"log"

	"github.com/EwanGreer/todo-cli/database"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var textColor = lipgloss.Color("#FAFAFA")

var (
	mainStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 2, 0, 1)
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(textColor).Padding(0, 1)
	itemStyle   = lipgloss.NewStyle().Foreground(textColor)
)

type model struct {
	choices []database.Task
	db      *database.Database
	cursor  int
}

func initialModel(db *database.Database) model {
	var tasks []database.Task
	tx := db.Find(&tasks)
	if tx.Error != nil {
		log.Fatal(tx.Error)
	}

	return model{
		choices: tasks,
		db:      db,
		cursor:  0,
	}
}

func (m model) Init() tea.Cmd {
	// TODO: should I be saving state here every loop, or does this only load on startup?
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var tasks []database.Task
	tx := m.db.Find(&tasks)
	if tx.Error != nil {
		log.Fatal(tx.Error)
	}
	m.choices = tasks

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			for _, task := range m.choices {
				tx := m.db.DB.Save(&task)
				if tx.Error != nil {
					continue
				}
			}

			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			if m.choices[m.cursor].Done {
				m.choices[m.cursor].Done = false
			} else {
				m.choices[m.cursor].Done = true
			}
			m.db.Save(&m.choices[m.cursor])
		}
	}

	return m, nil
}

func (m model) View() string {
	header := headerStyle.Render("Tasks:")

	var items []string
	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := " "
		if choice.Done {
			checked = "x"
		}

		// Render each item.
		item := itemStyle.Render(fmt.Sprintf("%s [%s] %s", cursor, checked, choice.Name))
		items = append(items, item)
	}

	instructions := "Press `q` to quit."

	view := lipgloss.JoinVertical(lipgloss.Left, header, lipgloss.JoinVertical(lipgloss.Left, items...), instructions)

	return mainStyle.Render(view)
}
