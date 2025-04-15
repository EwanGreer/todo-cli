package main

import (
	"fmt"
	"sort"

	"github.com/EwanGreer/todo-cli/database"
	"github.com/EwanGreer/todo-cli/internal/mode"
	"github.com/EwanGreer/todo-cli/internal/status"
	"github.com/charmbracelet/lipgloss"
)

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

			itemText := fmt.Sprintf("%s %s", cursor, list)
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

			itemText := fmt.Sprintf("%s [%s] %s", cursor, checked, choice)
			var item string
			if m.containers[containerTasks].cursor == i {
				item = selectedItemStyle.Render(itemText)
			} else {
				item = itemStyle.Render(itemText)
			}
			items = append(items, item)
		}

		var listView string
		if m.activeContainer == containerLists {
			highlightStyle := headerStyle.Foreground(lipgloss.Color("#FF5FAF"))
			listView = lipgloss.JoinVertical(
				lipgloss.Left,
				highlightStyle.Render("Lists:"),
				lipgloss.JoinVertical(lipgloss.Left, lists...),
				"",
			)
		} else {
			listView = lipgloss.JoinVertical(
				lipgloss.Left,
				headerStyle.Render("Lists:"),
				lipgloss.JoinVertical(lipgloss.Left, lists...),
				"",
			)
		}

		instructions := "Press `q` to quit | Press `a` to add a new todo | Press `d` to delete a todo"
		var taskView string
		if m.activeContainer == containerTasks {
			highlightStyle := headerStyle.Foreground(lipgloss.Color("#FF5FAF"))
			taskView = lipgloss.JoinVertical(
				lipgloss.Left,
				highlightStyle.Render("Tasks:"),
				lipgloss.JoinVertical(lipgloss.Left, items...),
				instructions,
			)
		} else {
			taskView = lipgloss.JoinVertical(
				lipgloss.Left,
				headerStyle.Render("Tasks:"),
				lipgloss.JoinVertical(lipgloss.Left, items...),
				instructions,
			)
		}

		return m.float(listView, taskView)
	case mode.ModeAdd:
		var inputView string
		switch m.activeContainer {
		case containerLists:
			inputView = m.addListTi.View()
		case containerTasks:
			inputView = m.addTaskTi.View()
		}
		view := fmt.Sprintf("Add New %s:\n\n", m.activeContainer)
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
