package botbuilder

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/durandj/wheatley/botbuilder/notifier"
)

var (
	statusImages = map[notifier.Status]string{
		notifier.StatusDebug: "🚧",
		notifier.StatusInfo:  "ℹ️",
		notifier.StatusWarn:  "⚠️",
		notifier.StatusError: "💥",
	}
)

// Bot models an assistant bot.
type Bot struct {
	name      string
	ctx       context.Context
	logger    *zap.SugaredLogger
	notifiers []notifier.Notifier
}

// NewBot creates a new bot instance.
func NewBot(name string, notifiers []notifier.Notifier) (*Bot, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("Unable to create logger for %s: %v", name, err)
	}

	if len(notifiers) == 0 {
		logger.Warn("No notifiers were specified, no notifications will be sent")
	}

	bot := Bot{
		name:      name,
		ctx:       context.Background(),
		logger:    logger.Sugar(),
		notifiers: notifiers,
	}

	return &bot, nil
}

// Start starts the bot.
func (bot *Bot) Start() {
	bot.logger.Infof("Starting bot %s", bot.name)
	bot.notifyDebug("Starting bot", "Bot is being started up")

	signals := make(chan os.Signal, 1)

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	<-signals

	bot.notifyDebug("Stopping bot", "Bot is being shutdown")
}

func (bot *Bot) notify(status notifier.Status, title string, body string) {
	for _, n := range bot.notifiers {
		err := n.SendNotification(
			bot.ctx,
			bot.name,
			status,
			fmt.Sprintf("%s %s", statusImages[status], title),
			body,
		)

		if err != nil {
			bot.logger.Errorf("Unable to send notification: %v", err)
		}
	}
}

func (bot *Bot) notifyDebug(title string, body string) {
	bot.notify(notifier.StatusDebug, title, body)
}

// func (bot *Bot) notifyInfo(title string, body string) {
// 	bot.notify(notifier.StatusInfo, title, body)
// }

// func (bot *Bot) notifyWarn(title string, body string) {
// 	bot.notify(notifier.StatusWarn, title, body)
// }

// func (bot *Bot) notifyError(title string, body string) {
// 	bot.notify(notifier.StatusError, title, body)
// }
