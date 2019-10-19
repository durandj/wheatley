package botbuilder

import (
	"fmt"
)

// TaskPriority is the priority of the task.
type TaskPriority int

// NewTaskPriority creates a new priority instance from a string.
func NewTaskPriority(priority string) TaskPriority {
	result, ok := stringToPriority[priority]
	if !ok {
		return NormalPriority
	}

	return result
}

func (priority TaskPriority) String() string {
	return priorityToString[priority]
}

const (
	// LowPriority tasks are less important and can be shed.
	LowPriority TaskPriority = -1

	// NormalPriority tasks have standard priority and won't be shed.
	NormalPriority TaskPriority = 0

	// HighPriority tasks take precedence over most other tasks and
	// won't be shed.
	HighPriority TaskPriority = 1
)

var (
	priorityToString = map[TaskPriority]string{
		LowPriority:    "low",
		NormalPriority: "normal",
		HighPriority:   "high",
	}

	stringToPriority = map[string]TaskPriority{
		"low":    LowPriority,
		"normal": NormalPriority,
		"high":   HighPriority,
	}
)

// Task is a unit of work that can be scheduled and run.
type Task struct {
	Name     string
	Priority TaskPriority
	Handler  func()
}

func (task Task) String() string {
	return fmt.Sprintf("Task{name: '%s'}", task.Name)
}
