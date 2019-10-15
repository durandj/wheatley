package notifier

import (
	"context"
	"fmt"

	"github.com/durandj/go-pushbullet"

	"github.com/durandj/wheatley/botbuilder/notifier"
)

// PushBulletNotifier is a notifier that sends notifications to
// PushBullet.
type PushBulletNotifier struct {
	pushBulletClient *pushbullet.Client
}

// NewPushBulletNotifier creates a notifier for sending to PushBullet.
func NewPushBulletNotifier(apiKey string) notifier.Notifier {
	return PushBulletNotifier{
		pushBulletClient: pushbullet.New(apiKey),
	}
}

// SendNotification sends a noification to PushBullet.
func (notifier PushBulletNotifier) SendNotification(
	ctx context.Context,
	botName string,
	status notifier.Status,
	title string,
	body string,
) error {
	return notifier.pushBulletClient.PushNoteWithContext(
		ctx,
		pushbullet.AllDevices,
		fmt.Sprintf("%s %s: %s", botName, status, title),
		body,
	)
}
