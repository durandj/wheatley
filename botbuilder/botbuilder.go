package botbuilder

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/robfig/cron/v3"
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
	name              string
	ctx               context.Context
	logger            *zap.SugaredLogger
	notifiers         []notifier.Notifier
	notificationLevel notifier.Status
	cronScheduler     *cron.Cron
	workQueue         *workQueue
	taskLock          *sync.RWMutex
	taskNotifier      *sync.Cond
}

// BotOpts are options used to configure a newly created bot.
type BotOpts struct {
	// Notifiers are things used to send notifications about things
	// that have happened.
	Notifiers []notifier.Notifier

	// NotificationLevel is the minimum level of notifications that
	// are allowed to be sent at a global level.
	NotificationLevel notifier.Status
}

// NewBot creates a new bot instance.
func NewBot(name string, opts BotOpts) (*Bot, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("Unable to create logger for %s: %v", name, err)
	}

	if len(opts.Notifiers) == 0 {
		logger.Warn("No notifiers were specified, no notifications will be sent")
	}

	taskLock := sync.RWMutex{}

	bot := Bot{
		name:              name,
		ctx:               context.Background(),
		logger:            logger.Sugar(),
		notifiers:         opts.Notifiers,
		notificationLevel: opts.NotificationLevel,
		cronScheduler:     cron.New(),
		workQueue:         newWorkQueue(),
		taskLock:          &taskLock,
		taskNotifier:      sync.NewCond(taskLock.RLocker()),
	}

	return &bot, nil
}

// Start starts the bot.
func (bot *Bot) Start() {
	bot.logger.Infof("Starting bot %s", bot.name)
	bot.notifyDebug("Starting bot", "Bot is being started up")

	shouldStop := false

	bot.taskLock.RLock()

	taskWaitGroup := sync.WaitGroup{}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signals

		shouldStop = true
		bot.taskNotifier.Signal()
	}()

	bot.cronScheduler.Start()
	for !shouldStop {
		for !shouldStop && bot.workQueue.IsEmpty() {
			bot.taskNotifier.Wait()
		}

		if shouldStop {
			break
		}

		task := bot.workQueue.Pop()
		taskWaitGroup.Add(1)
		go func() {
			defer taskWaitGroup.Done()

			task.Handler()
		}()
	}
	bot.cronScheduler.Stop()

	bot.logger.Debug("Waiting for running tasks to complete")
	taskWaitGroup.Wait()

	bot.logger.Infof("Stopping bot %s", bot.name)
	bot.notifyDebug("Stopping bot", "Bot is being shutdown")
}

// ScheduleTask schedules a task to run on a regular basis.
func (bot *Bot) ScheduleTask(schedule string, task Task) {
	_, err := bot.cronScheduler.AddFunc(schedule, func() {
		bot.logger.Infof("Queuing task: %v", task)

		bot.workQueue.Push(task)
		bot.taskNotifier.Signal()
	})

	if err != nil {
		bot.notifyError(
			"Unable schedule task",
			fmt.Sprintf("%s was unable to schedule '%s' task", bot.name, task.Name),
		)
	}
}

func (bot *Bot) notify(status notifier.Status, title string, body string) {
	if status < bot.notificationLevel {
		return
	}

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

func (bot *Bot) notifyError(title string, body string) {
	bot.notify(notifier.StatusError, title, body)
}
