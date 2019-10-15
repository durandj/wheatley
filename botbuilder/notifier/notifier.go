package notifier

import (
	"context"
)

// Status is the notification level.
type Status int

func (status Status) String() string {
	switch status {
	case StatusDebug:
		return "debug"

	case StatusInfo:
		return "info"

	case StatusWarn:
		return "warn"

	case StatusError:
		return "error"

	default:
		return "unknown"
	}
}

const (
	// StatusDebug is used to send debug notifications
	StatusDebug Status = -1

	// StatusInfo is used to send information notifications
	StatusInfo Status = 0

	// StatusWarn is used to send warning notifications
	StatusWarn Status = 1

	// StatusError is used to send error notifications
	StatusError Status = 2
)

// Notifier sends a notification of some change or event to the system
// admin/owner.
type Notifier interface {
	SendNotification(
		ctx context.Context,
		botName string,
		status Status,
		title string,
		body string,
	) error
}
