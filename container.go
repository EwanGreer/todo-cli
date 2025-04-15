package main

import (
	"log"

	"github.com/EwanGreer/todo-cli/database"
)

type Container uint

const (
	containerLists Container = iota
	containerTasks
)

func (c Container) String() string {
	if c == containerLists {
		return "List"
	}
	if c == containerTasks {
		return "Task"
	}

	return ""
}

type ContainerData struct {
	items  []Item
	cursor int
}

func NewContainer(db *database.Repository) map[Container]*ContainerData {
	var lists []database.List
	tx := db.Find(&lists)
	if tx.Error != nil {
		log.Fatal(tx.Error)
	}

	if len(lists) == 0 {
		lists = append(lists, database.List{
			Name: "Default",
		})
		db.Create(&lists)
	}

	tasks := db.FindTasksForList(&lists[0])
	if tx.Error != nil {
		log.Fatal(tx.Error)
	}

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
		tasksContainer.items = append(tasksContainer.items, task)
	}

	return map[Container]*ContainerData{
		containerLists: listsContainer,
		containerTasks: tasksContainer,
	}
}

func (c Container) CurrentItem(m *model) Item {
	return m.containers[c].items[m.containers[c].cursor]
}
