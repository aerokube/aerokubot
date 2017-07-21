package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"flag"
	"os"
	"fmt"
	"log"
	"strings"
)

var (
	token       string
	version     bool
	debug		bool
	gitRevision string = "HEAD"
	buildStamp  string = "unknown"
)

func init() {
	flag.StringVar(&token, "token", "", "Telegram bot token (required)")
	flag.BoolVar(&version, "version", false, "Show version and exit")
	flag.BoolVar(&debug, "debug", false, "Run in debug mode (will print all req/resp)")
	flag.Parse()


	if flag.NFlag() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	if version {
		showVersion()
		os.Exit(0)
	}
}

func showVersion() {
	fmt.Printf("Git Revision: %s\n", gitRevision)
	fmt.Printf("UTC Build Time: %s\n", buildStamp)
}

const welcome = `
Welcome to Aerokube chat! We can help on english 🇬🇧, так же как и по-русски 🇷🇺!

If you have trouble, provide your environment and versions for aerokube tools first!

Есть проблемы? Начни вопрос с окружения и используемой версии для тулчейна aerokube!
`

func main() {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(fmt.Errorf("[BOT_CREATION_FAIL] [%v]", err))
	}

	bot.Debug = debug

	log.Printf("Authorized on account %s, debug mode: %v", bot.Self.UserName, debug)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := bot.GetUpdatesChan(u)

	for update := range updates {

		if update.Message == nil {
			continue
		}

		if update.Message.Chat.IsGroup() {
			if update.Message.NewChatMembers != nil {
				newu := []string{}

				for _, user := range *update.Message.NewChatMembers {
					newu = append(newu, "@" + user.UserName)
				}

				ucall := strings.Join(newu, " ")


				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Hey, %s\n%s", ucall, welcome))
				bot.Send(msg)
			}
		}
	}
}
