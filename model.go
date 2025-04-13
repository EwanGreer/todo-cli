package main

import (
	"fmt"
	"log"
	"slices"
	"sort"

	"github.com/EwanGreer/todo-cli/database"
	"github.com/EwanGreer/todo-cli/internal/mode"
	"github.com/EwanGreer/todo-cli/internal/status"
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

func (m *model) CurrentListItem() Item {
	c := m.containers[m.activeContainer]
	return c.items[c.cursor]
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

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	case tea.WindowSizeMsg:
		m.updateWindowSize(msg)
		return m, nil
	}
	return m, nil
}

func (m *model) View() string {
	switch m.mode {
	case mode.ModeList:
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
	case mode.ModeAdd:
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
	m.mode = mode.ModeAdd
	m.addListTi.Focus()
}

func (m *model) addTask() {
	m.mode = mode.ModeAdd
	m.addTaskTi.Focus()
}

func (m *model) updateWindowSize(msg tea.WindowSizeMsg) {
	m.width = msg.Width
	m.height = msg.Height
}

type (
	MsgError       string
	MsgTaskCreated string
)

func (m *model) addTaskCmd() tea.Cmd {
	return func() tea.Msg {
		input := m.addTaskTi.Value()
		if input == "" {
			return nil
		}

		selectedList := m.CurrentList()
		task := database.NewTask(input, "", status.Ready, selectedList.ID)

		tx := m.db.DB.Save(&task)
		if tx.Error != nil {
			return MsgError(tx.Error.Error())
		}

		m.containers[containerTasks].items = append(m.containers[containerTasks].items, task)
		m.mode = mode.ModeList

		return MsgTaskCreated(fmt.Sprintf("Task %s created", task.Name))
	}
}

func (m *model) addListCmd() tea.Cmd {
	return func() tea.Msg {
		input := m.addListTi.Value()
		if input == "" {
			return nil
		}
		m.addListTi.Reset()

		list := database.NewList(input)

		tx := m.db.DB.Save(list)
		if tx.Error != nil {
			log.Println(tx.Error)
			return MsgError(tx.Error.Error())
		}

		m.containers[containerLists].items = append(m.containers[containerLists].items, list)
		m.mode = mode.ModeList

		return nil
	}
}

func (m *model) CurrentList() *database.List {
	if containerData, ok := m.containers[containerLists]; ok {
		if containerData.cursor < len(containerData.items) {
			if list, ok := containerData.items[containerData.cursor].(*database.List); ok {
				return list
			}
		}
	}
	return nil
}

func (m *model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.mode == mode.ModeAdd {
		return m.handleAddModeKey(msg)
	}
	return m.handleListModeKey(msg)
}

func (m *model) handleAddModeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		switch m.activeContainer {
		case containerLists:
			return m, m.addListCmd()
		case containerTasks:
			return m, m.addTaskCmd()
		}
	case "ctrl+c", "esc":
		m.mode = mode.ModeList
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

func (m *model) handleListModeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key := msg.String(); key {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "up", "k":
		m.handleUpKey()
		return m, nil

	case "down", "j":
		m.handleDownKey()
		return m, nil

	case "h", "l":
		m.toggleActiveContainer()
		return m, nil

	case "enter", " ", "x":
		m.toggleTaskStatus()
		return m, nil

	case "a":
		switch m.activeContainer {
		case containerLists:
			m.addList()
		case containerTasks:
			m.addTask()
		}
		return m, nil

	case "d":
		switch m.activeContainer {
		case containerLists:
			return m, m.deleteListCmd()
		case containerTasks:
			return m, m.deleteTaskCmd()
		}
	}
	return m, nil
}

func (m *model) handleUpKey() {
	switch m.activeContainer {
	case containerLists:
		m.decementCursor()
		m.updateTasksFromCurrentList()
	case containerTasks:
		m.decementCursor()
	}
}

func (m *model) handleDownKey() {
	switch m.activeContainer {
	case containerLists:
		m.incrementCursor()
		m.updateTasksFromCurrentList()
	case containerTasks:
		m.incrementCursor()
	}
}

func (m *model) updateTasksFromCurrentList() {
	tasks := m.db.FindTasksForList(m.CurrentList())
	items := make([]Item, len(tasks))
	for i, t := range tasks {
		items[i] = t
	}

	m.containers[containerTasks].items = items
}

func (m *model) toggleActiveContainer() {
	if m.activeContainer == containerLists {
		m.activeContainer = containerTasks
	} else {
		m.activeContainer = containerLists
	}
}

func (m *model) toggleTaskStatus() {
	task := m.CurrentTask()
	if task.Status.Is(status.Done) {
		task.Status = status.InProgress
	} else {
		task.Status = status.Done
	}
	m.db.Save(task)
}

func (m *model) deleteListCmd() tea.Cmd {
	return func() tea.Msg {
		if m.CurrentList().Name == "Default" {
			return nil
		}

		cursor := m.containers[containerLists].cursor
		if cursor < 0 || cursor >= len(m.containers[containerLists].items) {
			return nil
		}

		list, ok := m.containers[containerLists].items[cursor].(*database.List)
		if !ok || list == nil {
			return nil
		}

		tx := m.db.Delete(list)
		if tx.Error != nil {
			return MsgError(tx.Error.Error())
		}

		m.containers[containerLists].items = removeItem(m.containers[containerLists].items, cursor)
		if m.containers[containerLists].cursor >= len(m.containers[containerLists].items) && len(m.containers[containerLists].items) > 0 {
			m.containers[containerLists].cursor = len(m.containers[containerLists].items) - 1
		}

		if len(m.containers[containerLists].items) > 0 {
			selectedList := m.CurrentListItem().(*database.List)
			tasks := m.db.FindTasksForList(selectedList)
			items := make([]Item, len(tasks))
			for i, t := range tasks {
				items[i] = t
			}
			m.containers[containerTasks].items = items
		} else {
			m.containers[containerTasks].items = []Item{}
		}

		return "yay"
	}
}

func (m *model) deleteTaskCmd() tea.Cmd {
	return func() tea.Msg {
		cursor := m.containers[containerTasks].cursor
		if cursor < 0 || cursor >= len(m.containers[containerTasks].items) {
			return nil
		}

		task, ok := m.containers[containerTasks].items[cursor].(*database.Task)
		if !ok || task == nil {
			return nil
		}

		tx := m.db.Delete(task)
		if tx.Error != nil {
			return MsgError(tx.Error.Error())
		}

		m.containers[containerTasks].items = removeItem(m.containers[containerTasks].items, cursor)
		if m.containers[containerTasks].cursor >= len(m.containers[containerTasks].items) && len(m.containers[containerTasks].items) > 0 {
			m.containers[containerTasks].cursor = len(m.containers[containerTasks].items) - 1
		}

		return "yay"
	}
}

func removeItem(slice []Item, index int) []Item {
	return slices.Delete(slice, index, index+1)
}

func (m *model) CurrentTask() *database.Task {
	return m.containers[containerTasks].items[m.containers[containerTasks].cursor].(*database.Task)
}
