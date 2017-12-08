package main

import (
	"fmt"
	"log"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jmoiron/sqlx"
	_ "github.com/kshvakov/clickhouse"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
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

func selectFromClickhouse(connect *sqlx.DB, kind, method, date, queryType string) (int, error) {
	fmt.Println(date)
	var query string
	switch queryType {
	case "date":
		query = fmt.Sprintf("SELECT count(*) as count FROM results where method = '%v' and kind = '%v' "+
			"and created_date = toDate('%v') and namespace !='default'", method, kind, date)
	case "last":
		query = fmt.Sprintf("SELECT count(*) as count FROM results where method = '%v' "+
			"and kind = '%v' and created_date > toDate('%v') and namespace !='default'", method, kind, date)
	}

	result, err := connect.Query(query)
	if err != nil {
		return 0, fmt.Errorf("Can't select from clickhouse - %v\n", err)
	}
	count := 0
	for result.Next() {
		result.Scan(&count)
	}
	fmt.Println(count)
	return count, nil
}

type cmdHandlerFunc func(update tgbotapi.Update) error

var handlers = make(map[string]cmdHandlerFunc)

type services struct {
	clickhouse *sqlx.DB
	kube       *kubernetes.Clientset
	bot        *tgbotapi.BotAPI
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
}

func initServices() {
	config, err := clientcmd.BuildConfigFromFlags("", "admin.conf")
	if err != nil {
		fmt.Errorf("Can't connect to kubernetes - %v\n", err)
		panic(err)
	}

	// create the clientset
	svc.kube, err = kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Errorf("Can't create clientset for kubernetes - %v\n", err)
		panic(err)
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
	svc.bot.Debug = true
}

func init() {
	conf = OpenConfig("config.json")
	initHandlers()
	initServices()
}

func main() {
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

			fn, ok := handlers[update.Message.Command()]
			if !ok {
				svc.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "I don't know that command"))
			}
			if err := fn(update); err != nil {
				svc.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Error: "+err.Error()))
			}

		}

	}
}
