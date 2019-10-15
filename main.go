package main

import (
	"fmt"
	"os"

	"github.com/gobuffalo/envy"
	"github.com/logrusorgru/aurora"

	"github.com/durandj/wheatley/botbuilder"
	"github.com/durandj/wheatley/botbuilder/notifier"
	wheatleyNotifier "github.com/durandj/wheatley/notifier"
)

func main() {
	loadDotEnv()

	pushBulletAPIToken, err := envy.MustGet("PUSHBULLET_API_TOKEN")
	if err != nil {
		fmt.Println("Missing PushBullet API token, set PUSHBULLET_API_TOKEN")
		os.Exit(1)
	}

	notifiers := []notifier.Notifier{
		wheatleyNotifier.NewPushBulletNotifier(pushBulletAPIToken),
	}

	bot, err := botbuilder.NewBot("wheatley", notifiers)
	if err != nil {
		fmt.Println(
			aurora.Sprintf(aurora.Red("Unable to create bot instance: %v"), err),
		)
		os.Exit(1)
	}

	bot.Start()
}

func loadDotEnv() {
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		return
	}

	if err := envy.Load(".env"); err != nil {
		fmt.Printf("Unable to load .env: %v", err)
	}
}
