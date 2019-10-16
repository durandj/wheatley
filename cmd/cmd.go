package cmd

import (
	"fmt"
	"os"

	"github.com/gobuffalo/envy"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"

	"github.com/durandj/wheatley/botbuilder"
	"github.com/durandj/wheatley/botbuilder/notifier"
	wheatleyNotifier "github.com/durandj/wheatley/notifier"
)

var (
	notificationLevel string
)

var rootCmd = &cobra.Command{
	Use:     "wheatley",
	Short:   "Wheatley, the useless computer assistant",
	Version: "0.0.0",
	PreRun: func(cmd *cobra.Command, args []string) {
		if !isNotificationLevelValid(notificationLevel) {
			fmt.Println("Invalid notification level")
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		loadDotEnv()

		pushBulletAPIToken, err := envy.MustGet("PUSHBULLET_API_TOKEN")
		if err != nil {
			fmt.Println("Missing PushBullet API token, set PUSHBULLET_API_TOKEN")
			os.Exit(1)
		}

		notifiers := []notifier.Notifier{
			wheatleyNotifier.NewPushBulletNotifier(pushBulletAPIToken),
		}

		opts := botbuilder.BotOpts{
			Notifiers:         notifiers,
			NotificationLevel: notifier.NewStatus(notificationLevel),
		}
		bot, err := botbuilder.NewBot("wheatley", opts)
		if err != nil {
			fmt.Println(
				aurora.Sprintf(aurora.Red("Unable to create bot instance: %v"), err),
			)
			os.Exit(1)
		}

		bot.Start()
	},
}

func isNotificationLevelValid(notificationLevel string) bool {
	switch notificationLevel {
	case "debug":
		return true
	case "info":
		return true
	case "warn":
		return true
	case "error":
		return true
	default:
		return false
	}
}

func loadDotEnv() {
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		return
	}

	if err := envy.Load(".env"); err != nil {
		fmt.Printf("Unable to load .env: %v", err)
	}
}

func init() {
	flags := rootCmd.Flags()
	flags.StringVar(
		&notificationLevel,
		"notification-level",
		notifier.StatusInfo.String(),
		"The threshold before sending a notification. Valid valuas: debug, info, warn, error",
	)
}

// Execute runs the bot command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error running bot: %s\n", err)
		os.Exit(1)
	}
}
