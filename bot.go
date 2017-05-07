/*
lumber_bot is a Telegram bot that uses the Reddit API to share visceral images of Earth, Space, Cities, Architecture, and more with your group.

Add @kungfu_kenney_bot to your group to experience it today!

Your Own Bot

This package allows you to setup your own bot running locally or on a remote server to achieve the same goal.

1. First, register a simple, script Reddit app to get your client id and secret (https://github.com/reddit/reddit/wiki/OAuth2-Quick-Start-Example#first-steps)

2. Next, create a Telegram bot (https://core.telegram.org/bots#6-botfather)

3. Fill in etc/default-serverconf.yaml with the required values

5. Make any changes to etc/default-subreddits.yaml that you would like

6. Finally, run the bot
 lumber_bot -serverconf etc/serverconf -subredditconf etc/subredditconf

The bot will serve a random, highly upvoted, image from the past week from the list of supplied subreddits when someone uses the "/get" command
 */
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"lumber_bot/reddit"
	"lumber_bot/telegram"
)

const USER_AGENT = "KungFuKennyBot/0.9 20170501"

type Configuration struct {
	UserName     string `yaml:"username"`
	Password     string `yaml:"password"`
	ClientId     string `yaml:"clientid"`
	ClientSecret string `yaml:"clientsecret"`
	BotToken     string `yaml:"bottoken"`
	ServerCert   string `yaml:"servercert"`
	Remote       string `yaml:"remote"`
}

type Bot struct {
	Reddit     *reddit.Reddit
	Subreddits *reddit.Subreddits
	Telegram   *telegram.Telegram
	Conf       *Configuration
}

func NewConfiguration(filename *string) (*Configuration, error) {
	filepath, _ := filepath.Abs(*filename)
	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not read file %s: %s", filepath, err.Error()))
	}

	var conf Configuration
	err = yaml.Unmarshal(file, &conf)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not parse yaml: %s", err.Error()))
	}

	return &conf, nil
}

func main() {
	var (
		server    = flag.String("serverconf", "", "path to server configuration")
		subreddit = flag.String("subredditconf", "", "path to subreddit configuration")
	)
	flag.Parse()

	if *server == "" {
		log.Panicf("Server configuration needs to be supplied")
	}

	if *subreddit == "" {
		log.Panicf("Subreddit configuration needs to be supplied")
	}

	conf, err := NewConfiguration(server)
	if err != nil {
		log.Panicf("Could not read config: %s", err.Error())
	}

	subreddits, err := reddit.NewSubreddits(subreddit)
	if err != nil {
		log.Panicf("Could not read subreddits: %s", err.Error())
	}

	reddit, err := reddit.New(conf.UserName, conf.Password, conf.ClientId, conf.ClientSecret, USER_AGENT)
	if err != nil {
		log.Panicf("Could not initialize reddit client: %s", err.Error())
	}

	telegram := telegram.New(conf.BotToken, USER_AGENT)

	bot := &Bot{
		Telegram:   telegram,
		Reddit:     reddit,
		Subreddits: subreddits,
		Conf:       conf,
	}
	rand.Seed(time.Now().UTC().UnixNano())

	if conf.Remote != "" && conf.ServerCert != "" {
		url := fmt.Sprintf("https://%s/%s", conf.Remote, conf.BotToken)
		certfile, _ := filepath.Abs(conf.ServerCert)
		log.Printf("Registering webhook for %s using %s", url, certfile)

		err = telegram.SetWebhook(url, certfile)
		if err != nil {
			log.Panicf("Could not register webhook for %s: %s", url, err.Error())
		}
		log.Printf("Registered webhook")

		mux := http.NewServeMux()
		mux.HandleFunc(fmt.Sprintf("/%s", conf.BotToken), bot.HandleRequest)
		mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
			fmt.Fprintf(w, "healthy")
		})

		log.Printf("Listening...")
		http.ListenAndServe(":5000", mux)
	} else {
		log.Printf("Polling for updates...")
		bot.Poll()
	}
}

func (b *Bot) HandleRequest(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("bot: http: error: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var update telegram.Update
	if err = json.Unmarshal([]byte(body), &update); err != nil {
		log.Printf("bot: Could not unmarshall request: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = b.Handle(update)
	if err != nil {
		log.Printf("bot: error handling update: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (b *Bot) Handle(update telegram.Update) error {
	command := strings.Split(update.Message.Text, " ")[0]
	if command != "/get" && command != "/get@kungfu_kenny_bot" {
		return nil
	}

	ChatId := update.Message.Chat.Id
	subreddit := b.Subreddits.PickRandom(ChatId)
	log.Printf("bot: Picked subreddit %s", subreddit)

	result, _ := b.Reddit.TopListings(subreddit, "week", 5)

	log.Printf("bot: Received %d results for subreddit %s", len(result.ResponseData.Listings), subreddit)

	listing := result.ResponseData.Listings[rand.Intn(len(result.ResponseData.Listings))].Listing
	message := fmt.Sprintf("/r/%s: %s %s", listing.Subreddit, listing.Title, listing.Url)
	ReplyMessageId := update.Message.MessageId

	log.Printf("bot: Replying to message %d in chat %d with: %s", ReplyMessageId, ChatId, message)
	err := b.Telegram.SendMessage(ChatId, message, ReplyMessageId)
	if err != nil {
		log.Printf("bot: Could not respond to chat:%d message:%d :%s", ChatId, ReplyMessageId, err.Error())
		return err
	}
	return nil
}

func (b *Bot) Poll() {
	offset := 0
	for {
		updates, err := b.Telegram.GetUpdates(offset)
		if err != nil {
			log.Printf("bot: Received error fetching updates: %s", err.Error())
		}
		log.Printf("bot: Received %d updates", len(updates))

		for _, update := range updates {
			offset = update.UpdateId
			go b.Handle(update)
		}

		if offset != 0 {
			offset += 1
		}
		time.Sleep(250)
	}
}
