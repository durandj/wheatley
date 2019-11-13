package botbuilder

import (
	"fmt"
)

// Task is a unit of work that can be scheduled and run.
type Task struct {
	Name    string
	Handler func()
}

func (task Task) String() string {
	return fmt.Sprintf("Task{name: '%s'}", task.Name)
}
