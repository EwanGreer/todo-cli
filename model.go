package main

import (
	"fmt"
	"log"
	"sort"

	"github.com/EwanGreer/todo-cli/database"
	"github.com/EwanGreer/todo-cli/internal/status"
	"github.com/charmbracelet/bubbles/textinput"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Mode uint

type Container uint

const (
	modeList Mode = iota
	modeAdd
)

const (
	containerLists Container = iota
	containerTasks
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

type ContainerData struct {
	items  []Item
	cursor int
}

type model struct {
	containers      map[Container]*ContainerData
	db              *database.Repository
	activeContainer Container
	width           int
	height          int
	mode            Mode
	addTaskTi       textinput.Model
	addListTi       textinput.Model
}

func initialModel(db *database.Repository) *model {
	addTaskTi := textinput.New()
	addTaskTi.Placeholder = "Enter new todo..."
	addTaskTi.CharLimit = 100
	addTaskTi.Width = 30

	addListTi := textinput.New()
	addListTi.Placeholder = "Enter a new list..."
	addListTi.CharLimit = 100
	addListTi.Width = 30

	var lists []database.List
	tx := db.Find(&lists)
	if tx.Error != nil {
		log.Fatal(tx.Error)
	}

	var tasks []database.Task
	tx = db.Find(&tasks)
	if tx.Error != nil {
		log.Fatal(tx.Error)
	}

	// Initialize the containers
	listsContainer := &ContainerData{
		items:  make([]Item, 0),
		cursor: 0,
	}

	tasksContainer := &ContainerData{
		items:  make([]Item, 0),
		cursor: 0,
	}

	for _, list := range lists {
		listsContainer.items = append(listsContainer.items, &list)
	}

	for _, task := range tasks {
		tasksContainer.items = append(tasksContainer.items, &task)
	}

	return &model{
		containers: map[Container]*ContainerData{
			containerLists: listsContainer,
			containerTasks: tasksContainer,
		},
		db:              db,
		addTaskTi:       addTaskTi,
		mode:            modeList,
		activeContainer: containerLists,
		addListTi:       addListTi,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var tasks []database.Task
	selectedList := m.containers[containerLists].items[m.containers[containerLists].cursor].(*database.List)
	tx := m.db.Find(&tasks, "list_id = ?", selectedList.ID)
	if tx.Error != nil {
		log.Fatal(tx.Error)
	}

	newTasks := make([]Item, 0, len(tasks))
	for i := range tasks {
		newTasks = append(newTasks, &tasks[i])
	}
	m.containers[containerTasks].items = newTasks

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case modeAdd:
			switch msg.String() {
			case "enter":
				switch m.activeContainer {
				case containerLists:
					var list *database.List
					if input := m.addListTi.Value(); input != "" {
						list = database.NewList(input)
					}

					tx := m.db.DB.Save(&list)
					if tx.Error != nil {
						log.Println(tx.Error)
						return m, nil
					}

					m.containers[containerLists].items = append(m.containers[containerLists].items, list)
					m.mode = modeList

					return m, nil
				case containerTasks:
					var task *database.Task
					if input := m.addTaskTi.Value(); input != "" {
						selectedList := m.containers[containerLists].items[m.containers[containerLists].cursor].(*database.List)
						task = database.NewTask(input, "", status.Ready, selectedList.ID)
					}
					tx := m.db.DB.Save(&task)
					if tx.Error != nil {
						log.Println(tx.Error)
						return m, nil
					}
					m.containers[containerTasks].items = append(m.containers[containerTasks].items, task)
					m.mode = modeList
					return m, nil
				}
			case "ctrl+c", "esc":
				m.mode = modeList
				return m, nil
			}

			var cmd tea.Cmd
			switch m.activeContainer {
			case containerLists:
				m.addListTi, cmd = m.addListTi.Update(msg)
			case containerTasks:
				m.addTaskTi, cmd = m.addTaskTi.Update(msg)
			}
			return m, cmd
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			m.decementCursor()

			selectedList := m.containers[containerLists].items[m.containers[containerLists].cursor].(*database.List)
			tasks := m.db.FindTasksForList(selectedList)
			items := make([]Item, len(tasks))
			for i, t := range tasks {
				items[i] = t
			}
			m.containers[containerTasks].items = items
			return m, nil
		case "down", "j":
			m.incrementCursor()

			selectedList := m.containers[containerLists].items[m.containers[containerLists].cursor].(*database.List)
			tasks := m.db.FindTasksForList(selectedList)
			items := make([]Item, len(tasks))
			for i, t := range tasks {
				items[i] = t
			}
			m.containers[containerTasks].items = items
			return m, nil
		case "h", "l":
			if m.activeContainer == containerLists {
				m.activeContainer = containerTasks
			} else {
				m.activeContainer = containerLists
			}
		case "enter", " ", "x":
			task := m.containers[containerTasks].items[m.containers[containerTasks].cursor].(*database.Task)
			if task.Status.Is(status.Done) {
				task.Status = status.InProgress
			} else {
				task.Status = status.Done
			}
			m.db.Save(task)
		case "a":
			switch m.activeContainer {
			case containerLists:
				m.addList()
			case containerTasks:
				m.addTask()
			}
		}
	case tea.WindowSizeMsg:
		m.updateWindowSize(msg)
	}

	return m, nil
}

func (m model) View() string {
	switch m.mode {
	case modeList:
		sort.Slice(m.containers[containerTasks].items, func(i, j int) bool {
			return m.containers[containerTasks].items[i].(*database.Task).CreatedAt.Before(m.containers[containerTasks].items[j].(*database.Task).CreatedAt)
		})

		var lists []string
		for i, list := range m.containers[containerLists].items {
			cursor := " "
			if m.containers[containerLists].cursor == i {
				cursor = ">"
			}

			itemText := fmt.Sprintf("%s %s", cursor, list.String())
			var item string
			if m.containers[containerLists].cursor == i {
				item = selectedItemStyle.Render(itemText)
			} else {
				item = itemStyle.Render(itemText)
			}
			lists = append(lists, item)
		}

		var items []string
		for i, choice := range m.containers[containerTasks].items {
			cursor := " "
			if m.containers[containerTasks].cursor == i {
				cursor = ">"
			}

			checked := " "
			if choice.(*database.Task).Status.Is(status.Done) {
				checked = "x"
			}

			itemText := fmt.Sprintf("%s [%s] %s", cursor, checked, choice.String())
			var item string
			if m.containers[containerTasks].cursor == i {
				item = selectedItemStyle.Render(itemText)
			} else {
				item = itemStyle.Render(itemText)
			}
			items = append(items, item)
		}

		listView := lipgloss.JoinVertical(
			lipgloss.Left,
			headerStyle.Render("Lists:"),
			lipgloss.JoinVertical(lipgloss.Left, lists...),
			"",
		)

		instructions := "Press `q` to quit | Press `a` to add a new todo | Press `d` to delete a todo"

		taskView := lipgloss.JoinVertical(
			lipgloss.Left,
			headerStyle.Render("Tasks:"),
			lipgloss.JoinVertical(lipgloss.Left, items...),
			instructions,
		)

		return m.float(listView, taskView)
	case modeAdd:
		var inputView string
		switch m.activeContainer {
		case containerLists:
			inputView = m.addListTi.View()
		case containerTasks:
			inputView = m.addTaskTi.View()
		}
		view := "Add New TODO:\n\n"
		view += inputView + "\n\n"
		view += "Press Enter to confirm, Esc to cancel.\n"
		return m.float(view)
	}

	return "Unknown Mode"
}

// float accepts a veriadic string of "columns"
// each "column" will be given a border
func (m model) float(views ...string) string {
	var styledViews []string
	for _, view := range views {
		styledViews = append(styledViews, mainStyle.Render(view))
	}

	combinedView := lipgloss.JoinHorizontal(lipgloss.Center, styledViews...)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, mainStyle.Render(combinedView))
}

func (m *model) decementCursor() {
	if containerData, ok := m.containers[m.activeContainer]; ok && containerData.cursor > 0 {
		containerData.cursor--
	}
}

func (m *model) incrementCursor() {
	if containerData, ok := m.containers[m.activeContainer]; ok && containerData.cursor < len(containerData.items)-1 {
		containerData.cursor++
	}
}

func (m *model) addList() {
	m.mode = modeAdd
	m.addListTi.Focus()
}

func (m *model) addTask() {
	m.mode = modeAdd
	m.addTaskTi.Focus()
}

func (m *model) updateWindowSize(msg tea.WindowSizeMsg) {
	m.width = msg.Width
	m.height = msg.Height
}
