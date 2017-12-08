package main

import (
	"fmt"
	"strconv"

	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func helloHandler(update tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	msg.Text = "Стартуем, сегодня мы с тобой стартуем!"
	msg.ReplyMarkup = numericKeyboard
	svc.bot.Send(msg)

	return nil
}

func helpHandler(update tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	msg.Text = fmt.Sprintf("/date %s\n"+
		"/last 90\n"+
		"/nscount \n"+
		"/deploycount\n"+
		"/images", tbotDateExample)
	svc.bot.Send(msg)

	return nil
}

func lastHandler(update tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	last, err := strconv.ParseInt(update.Message.CommandArguments(), 0, 64)
	if err != nil {
		return fmt.Errorf("Wrong fomrat, use - %v", tbotDateExample)
	}
	date := time.Now().Add(-24 * time.Hour * time.Duration(last))
	method, kind := "create", "namespaces"
	nsCount, err := selectFromClickhouse(svc.clickhouse, kind, method, date.Format(chDateFormat), "last")
	if err != nil {
		return fmt.Errorf("Can't select ns from clickhouse")
	}
	kind = "deployments"
	deployCount, err := selectFromClickhouse(svc.clickhouse, kind, method, date.Format(chDateFormat), "last")
	if err != nil {
		return fmt.Errorf("Can't select deployments from clickhouse")
	}

	msg.Text = fmt.Sprintf("Начиная с  - %v \n"+
		"Количество созданных неймспейсов - %v\n"+
		"Количество созданный деплоев - %v\n", date.Format(chDateFormat), nsCount, deployCount)
	svc.bot.Send(msg)

	return nil
}

func dateHandler(update tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	date, err := time.Parse(tbotDateFormat, update.Message.CommandArguments())
	if err != nil {
		return fmt.Errorf("Wrong fomrat, use - %v", tbotDateExample)
	}
	method, kind := "create", "namespaces"
	nsCount, err := selectFromClickhouse(svc.clickhouse, kind, method, date.Format(chDateFormat), "date")
	if err != nil {
		return fmt.Errorf("Can't select ns from clickhouse")
	}
	kind = "deployments"
	deployCount, err := selectFromClickhouse(svc.clickhouse, kind, method, date.Format(chDateFormat), "date")
	if err != nil {
		return fmt.Errorf("Can't select deployments from clickhouse")
	}

	msg.Text = fmt.Sprintf("Начиная с  - %v \n"+
		"Количество созданных неймспейсов - %v\n"+
		"Количество созданный деплоев - %v\n", date.Format(chDateFormat), nsCount, deployCount)
	svc.bot.Send(msg)

	return nil
}

func nsCount(update tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	count, err := GetNsCount(svc.kube)
	if err != nil {
		return fmt.Errorf("Can't select ns count from kube")
	}
	msg.Text = fmt.Sprintf("Количество неймспейсов в кубе - %v\n", count)
	svc.bot.Send(msg)

	return nil
}

func deployCount(update tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	count, err := GetDeployCount(svc.kube)
	if err != nil {
		return fmt.Errorf("Can't select deployments count from kube")
	}
	msg.Text = fmt.Sprintf("Количество деплойментов в кубе - %v\n", count)
	svc.bot.Send(msg)

	return nil
}

func getImages(update tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	active, notActive, err := GetImages(svc.kube)
	if err != nil {
		return fmt.Errorf("Can't select deployments from kube")
	}
	activeString := "\n"
	notActiveString := "\n"
	for k, v := range active {
		activeString += fmt.Sprintf("%02d - %s \n", v, k)
	}

	for k, v := range notActive {
		notActiveString += fmt.Sprintf("%02d - %s \n", v, k)
	}
	msg.Text = fmt.Sprintf("Активные image: %v\n"+
		"Неактивные image:  %v\n", activeString, notActiveString)
	svc.bot.Send(msg)

	return nil
}
