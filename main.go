package main

import (
	"flag"
	"os"
	"fmt"
	"log"
	"strings"
	"net/http"
	"encoding/json"
	"bytes"
	"time"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/dustin/go-humanize"
)

var (
	token       string
	ghtoken     string
	version     bool
	debug       bool
	gitRevision string = "HEAD"
	buildStamp  string = "unknown"
)

func init() {
	flag.StringVar(&token, "token", "", "Telegram bot token (required)")
	flag.StringVar(&ghtoken, "github-token", "", "GitHub token with public read permissions")
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
Welcome to Aerokube chat! We can help in English 🇬🇧, так же как и по-русски 🇷🇺!

Having troubles? Please provide your environment and Aerokube tools versions!

Есть проблемы? Начни вопрос с окружения и используемой версии инструментов Aerokube!
`

type gql struct {
	Query string `json:"query"`
}

type result struct {
	Data map[string]repo `json:"data"`
}

type repo struct {
	Releases struct {
		Nodes []release `json:"nodes"`
	} `json:"releases"`
}

type release struct {
	Url         string `json:"url"`
	PublishedAt time.Time `json:"publishedAt"`
	Tag struct {
		Name string `json:"name"`
	} `json:"tag"`
}

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

		if update.Message.Chat.IsGroup() || update.Message.Chat.IsSuperGroup() {
			if update.Message.NewChatMembers != nil {
				newu := []string{}

				for _, user := range *update.Message.NewChatMembers {
					newu = append(newu, "@"+getUserName(user))
				}

				ucall := strings.Join(newu, " ")

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Hey, %s\n%s", ucall, welcome))
				bot.Send(msg)
			}
		}

		// COMMANDS
		if update.Message.IsCommand() {

			switch update.Message.Command() {
			case "releases":
				log.Println("Executing releases command")
				result := make(chan string)
				go releases(result)

				select {
				case msg := <-result:
					resp := tgbotapi.NewMessage(update.Message.Chat.ID, msg)
					resp.ReplyToMessageID = update.Message.MessageID
					resp.ParseMode = "markdown"
					bot.Send(resp)
				case <-time.After(10 * time.Second):
				}
			}
		}
	}
}

func getUserName(user tgbotapi.User) string {
	if user.UserName == "" {
		return user.FirstName
	}
	return user.UserName
}

func releases(msg chan<- string) {
	query := `
fragment release on Repository {
  releases(last: 1) {
    nodes {
      url
      publishedAt
      tag {
        name
      }
    }
  }
}

query repos {
  selenoid: repository(owner: "aerokube", name: "selenoid") {
    ...release
  }
  cm: repository(owner: "aerokube", name: "cm") {
    ...release
  }
  selenoid_ui: repository(owner: "aerokube", name: "selenoid-ui") {
    ...release
  }
  ggr: repository(owner: "aerokube", name: "ggr") {
    ...release
  }
}
`

	q, err := json.Marshal(gql{Query: query})
	if err != nil {
		log.Printf("Cant marshall query: %v\n", err)
		return
	}

	req, err := http.NewRequest(
		"POST",
		"https://api.github.com/graphql",
		bytes.NewReader(q),
	)
	if err != nil {
		log.Printf("Failed to create GH request: %v\n", err)
		return
	}
	req.Header.Add("Authorization", "Bearer "+ghtoken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Cannot fetch GH releases for aerokube: %v\n", err)
		return
	}
	defer resp.Body.Close()

	result := &result{}

	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		log.Printf("Cant unmarshal GH response: %v\n", err)
		return
	}

	repos := []string{}

	for name, repo := range result.Data {
		rel := repo.Releases.Nodes[0]

		repos = append(repos, fmt.Sprintf(
			"*%s*: [%s](%s) - %s",
			name,
			rel.Tag.Name,
			rel.Url,
			humanize.Time(rel.PublishedAt),
		))
	}

	msg <- strings.Join(repos, "\n")
}
