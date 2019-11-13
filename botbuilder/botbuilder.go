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

const (
	defaultWorkerCount = 32
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
	workerCount       int
	workQueue         chan Task
}

// BotOpts are options used to configure a newly created bot.
type BotOpts struct {
	// Notifiers are things used to send notifications about things
	// that have happened.
	Notifiers []notifier.Notifier

	// NotificationLevel is the minimum level of notifications that
	// are allowed to be sent at a global level.
	NotificationLevel notifier.Status

	// MaxConcurrency sets the maximum number of concurrently running
	// tasks. If unset a default concurrency amount is used.
	MaxConcurrency int
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

	if opts.MaxConcurrency < 0 {
		return nil, fmt.Errorf(
			"Invalid concurrency setting, must be non-negative but was %d",
			opts.MaxConcurrency,
		)
	} else if opts.MaxConcurrency == 0 {
		opts.MaxConcurrency = defaultWorkerCount
	}

	bot := Bot{
		name:              name,
		ctx:               context.Background(),
		logger:            logger.Sugar(),
		notifiers:         opts.Notifiers,
		notificationLevel: opts.NotificationLevel,
		cronScheduler:     cron.New(),
		workerCount:       opts.MaxConcurrency,
		workQueue:         make(chan Task),
	}

	return &bot, nil
}

// Start starts the bot.
func (bot *Bot) Start() {
	bot.logger.Infof("Starting bot %s", bot.name)
	bot.notifyDebug("Starting bot", "Bot is being started up")

	taskWaitGroup := sync.WaitGroup{}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	for index := 0; index < bot.workerCount; index++ {
		go func() {
			for task := range bot.workQueue {
				task.Handler()
			}

			taskWaitGroup.Done()
		}()
		taskWaitGroup.Add(1)
	}

	bot.cronScheduler.Start()
	<-signals
	bot.cronScheduler.Stop()
	close(bot.workQueue)

	bot.logger.Debug("Waiting for running tasks to complete")
	taskWaitGroup.Wait()

	bot.logger.Infof("Stopping bot %s", bot.name)
	bot.notifyDebug("Stopping bot", "Bot is being shutdown")
}

// ScheduleTask schedules a task to run on a regular basis.
func (bot *Bot) ScheduleTask(schedule string, task Task) {
	_, err := bot.cronScheduler.AddFunc(schedule, func() {
		bot.logger.Infof("Queuing task: %v", task)

		bot.workQueue <- task
	})

	if err != nil {
		bot.notifyError(
			"Unable schedule task",
			fmt.Sprintf("%s was unable to schedule '%s' task", bot.name, task.Name),
		)
		bot.logger.Errorf("Unable to schedule task '%s': %v", task, err)
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
