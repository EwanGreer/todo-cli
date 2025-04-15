package main

import (
	"fmt"

	"github.com/EwanGreer/todo-cli/database"
	"github.com/EwanGreer/todo-cli/internal/mode"
	"github.com/charmbracelet/bubbles/textinput"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	textColor         = lipgloss.Color("#FAFAFA")
	selectedItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5FAF"))

	mainStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 2, 0, 1)
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(textColor).Padding(0, 1)
	itemStyle   = lipgloss.NewStyle().Foreground(textColor)
)

type Item interface {
	fmt.Stringer
}

type model struct {
	containers      map[Container]*ContainerData
	db              *database.Repository
	activeContainer Container
	width           int
	height          int
	mode            mode.Mode
	addTaskTi       textinput.Model
	addListTi       textinput.Model
}

func initialModel(db *database.Repository) *model {
	addTaskTi := textinput.New()
	addTaskTi.Placeholder = "Enter new todo..."
	addTaskTi.CharLimit = 100
	addTaskTi.Width = 50

	addListTi := textinput.New()
	addListTi.Placeholder = "Enter a new list..."
	addListTi.CharLimit = 100
	addListTi.Width = 50

	return &model{
		containers:      NewContainer(db),
		db:              db,
		addTaskTi:       addTaskTi,
		mode:            mode.ModeList,
		activeContainer: containerLists,
		addListTi:       addListTi,
	}
}

func (m *model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *model) CurrentListItem() Item {
	c := m.containers[m.activeContainer]
	return c.items[c.cursor]
}
