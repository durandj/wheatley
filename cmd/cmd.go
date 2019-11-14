package cmd

import (
	"fmt"
	"os"
	"regexp"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/durandj/wheatley/botbuilder"
	"github.com/durandj/wheatley/botbuilder/notifier"
	wheatleyNotifier "github.com/durandj/wheatley/notifier"
)

const (
	pushBulletTokenKey = "pushbullet_token"
)

var (
	notificationLevelPattern = regexp.MustCompile("^debug|info|warn|error$")
	defaultConfigPath        = os.ExpandEnv("$HOME/.config/wheatley.yml")

	configFilePath    string
	maxWorkers        int
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
		setupConfig()

		if !viper.IsSet(pushBulletTokenKey) {
			fmt.Println("Missing PushBullet API token, set PUSHBULLET_API_TOKEN")
			os.Exit(1)
		}
		pushBulletToken := viper.GetString(pushBulletTokenKey)

		notifiers := []notifier.Notifier{
			wheatleyNotifier.NewPushBulletNotifier(pushBulletToken),
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

		bot.ScheduleTask(
			"* * * * *",
			botbuilder.Task{
				Name: "test",
				Handler: func() {
					fmt.Println("test")
				},
			},
		)
		for i := 1; i <= 500; i++ {
			bot.ScheduleTask(
				fmt.Sprintf("* * * * */%d", i%10+1),
				botbuilder.Task{
					Name: fmt.Sprintf("test%d", i),
					Handler: func() {
						fmt.Println("test")
					},
				},
			)
		}

		bot.Start()
	},
}

func isNotificationLevelValid(notificationLevel string) bool {
	return notificationLevelPattern.MatchString(notificationLevel)
}

func setupConfig() {
	viper.SetConfigFile("wheatley")
	viper.SetConfigFile("/etc/wheatley")
	viper.AddConfigPath("$HOME/.config")
	viper.SetEnvPrefix("wheatley")

	// The only possible error that can be returned is for no
	// arguments being passed. As we're providing an argument, the
	// return value isn't helpful.
	// nolint:errcheck,gosec
	viper.BindEnv(pushBulletTokenKey)

	if configFilePath != "" {
		viper.SetConfigFile(configFilePath)
	}
}

func init() {
	flags := rootCmd.Flags()

	flags.StringVar(
		&configFilePath,
		"config",
		defaultConfigPath,
		"The file path to load configuration from",
	)

	flags.StringVar(
		&notificationLevel,
		"notification-level",
		notifier.StatusInfo.String(),
		"The threshold before sending a notification. Valid valuas: debug, info, warn, error",
	)

	flags.IntVar(
		&maxWorkers,
		"max-concurrency",
		0,
		"The maximum concurrent tasks, 0 means infinite",
	)
}

// Execute runs the bot command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error running bot: %s\n", err)
		os.Exit(1)
	}
}
