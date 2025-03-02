package main

import (
	"fmt"
	"log"

	"github.com/EwanGreer/todo-cli/database"
	"github.com/charmbracelet/bubbles/textinput"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Mode int

const (
	modeList Mode = iota
	modeAdd
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
	width   int
	height  int
	adding  bool
	mode    Mode
	ti      textinput.Model
}

func initialModel(db *database.Database) model {
	ti := textinput.New()
	ti.Placeholder = "Enter new todo..."
	ti.CharLimit = 100
	ti.Width = 30

	var tasks []database.Task
	tx := db.Find(&tasks)
	if tx.Error != nil {
		log.Fatal(tx.Error)
	}

	return model{
		choices: tasks,
		db:      db,
		cursor:  0,
		ti:      ti,
		mode:    modeList,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
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
		switch m.mode {
		case modeList:
			switch msg.String() {
			case "ctrl+c", "q":
				for _, task := range m.choices {
					tx := m.db.DB.Save(&task)
					if tx.Error != nil {
						continue
					}
				}
				return m, tea.Quit
			case "a":
				m.mode = modeAdd
				m.ti.SetValue("")
				m.ti.Focus()
			}
		case modeAdd:
			switch msg.String() {
			case "enter":
				if input := m.ti.Value(); input != "" {
					m.choices = append(m.choices, database.Task{
						Name: input,
					})
				}

				for _, task := range m.choices {
					tx := m.db.DB.Save(&task)
					if tx.Error != nil {
						continue
					}
				}

				m.mode = modeList
				return m, nil
			case "esc":
				m.mode = modeList
				return m, nil
			}
			var cmd tea.Cmd
			m.ti, cmd = m.ti.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "ctrl+c", "q":

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
		case "a":
			m.adding = true
			m.ti.Focus()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m model) View() string {
	switch m.mode {
	case modeList:
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

			item := itemStyle.Render(fmt.Sprintf("%s [%s] %s", cursor, checked, choice.Name))
			items = append(items, item)
		}

		instructions := "Press `q` to quit. | "
		if m.adding {
			instructions += "New Todo: " + m.ti.View()
		} else {
			instructions += "Press 'a' to add a new todo."
		}

		view := lipgloss.JoinVertical(lipgloss.Left, header, lipgloss.JoinVertical(lipgloss.Left, items...), instructions)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, mainStyle.Render(view))
	case modeAdd:
		view := "Add New TODO:\n\n"
		view += m.ti.View() + "\n\n"
		view += "Press Enter to confirm, Esc to cancel.\n"

		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, mainStyle.Render(view))
	default:
		return "Unknown mode"
	}
}
