package main

import (
	"fmt"
	"log"

	"strconv"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jmoiron/sqlx"
	_ "github.com/kshvakov/clickhouse"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	tbotDateFormat  = "02.01.2006"
	chDateFormat    = "2006-01-02"
	tbotDateExample = "28.11.2017"
)

var numericKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("/images"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("/nscount"),
		tgbotapi.NewKeyboardButton("/deploycount"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("/last 7"),
		tgbotapi.NewKeyboardButton("/last 30"),
		tgbotapi.NewKeyboardButton("/last 90"),
	),
)

type cmdHandlerFunc func(update tgbotapi.Update) error

var handlers = make(map[string]cmdHandlerFunc)

type services struct {
	clickhouse *sqlx.DB
	kube       *kubernetes.Clientset
	bot        *tgbotapi.BotAPI
	postgres   *sqlx.DB
}

var svc services
var conf Config

func initHandlers() {
	handlers["start"] = helloHandler
	handlers["help"] = helpHandler
	handlers["last"] = lastHandler
	handlers["date"] = dateHandler
	handlers["nscount"] = nsCount
	handlers["deploycount"] = deployCount
	handlers["images"] = getImages
	handlers["info"] = getEmailInfo
}

func initServices() {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(fmt.Errorf("Can't connect to kubernetes - %v\n", err))
	}

	// create the clientset
	svc.kube, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(fmt.Errorf("Can't create clientset for kubernetes - %v\n", err))
	}

	svc.clickhouse, err = sqlx.Open("clickhouse",
		conf.Clickhouse.Url)
	if err != nil {
		log.Panic(err)
	}
	svc.clickhouse.Begin()

	svc.bot, err = tgbotapi.NewBotAPI(conf.TelegramBot.Token)
	if err != nil {
		log.Panic(err)
	}
	svc.bot.Debug = false

	svc.postgres, err = sqlx.Open("postgres",
		conf.Postgres.Url)
	if err != nil {
		log.Panic(err)
	}
	svc.postgres.Begin()
}

func main() {
	conf = OpenConfig()
	initHandlers()
	initServices()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := svc.bot.GetUpdatesChan(u)
	if err != nil {
		log.Panic(err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			if _, ok := conf.TelegramBot.Users[strconv.Itoa(update.Message.From.ID)]; !ok {
				continue
			}

			fn, ok := handlers[update.Message.Command()]
			if !ok {
				svc.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "I don't know that command"))
				continue
			}

			if err := fn(update); err != nil {
				svc.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Error: "+err.Error()))
			}
		}
	}
}
