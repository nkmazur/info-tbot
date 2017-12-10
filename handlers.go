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
		return fmt.Errorf("Wrong fomrat, use - %v\n", tbotDateExample)
	}
	date := time.Now().Add(-24 * time.Hour * time.Duration(last))
	method, kind := "create", "namespaces"
	nsCount, err := selectFromClickhouse(kind, method, date.Format(chDateFormat), "last")
	if err != nil {
		return fmt.Errorf("Can't select ns from clickhouse - %v\n", err)
	}
	kind = "deployments"
	deployCount, err := selectFromClickhouse(kind, method, date.Format(chDateFormat), "last")
	if err != nil {
		return fmt.Errorf("Can't select deployments from clickhouse - %v\n", err)
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
		return fmt.Errorf("Wrong fomrat, use - %v\n", tbotDateExample)
	}
	method, kind := "create", "namespaces"
	nsCount, err := selectFromClickhouse(kind, method, date.Format(chDateFormat), "date")
	if err != nil {
		return fmt.Errorf("Can't select ns from clickhouse - %v\n", err)
	}
	kind = "deployments"
	deployCount, err := selectFromClickhouse(kind, method, date.Format(chDateFormat), "date")
	if err != nil {
		return fmt.Errorf("Can't select deployments from clickhouse - %v\n", err)
	}

	msg.Text = fmt.Sprintf("Начиная с  - %v \n"+
		"Количество созданных неймспейсов - %v\n"+
		"Количество созданный деплоев - %v\n", date.Format(chDateFormat), nsCount, deployCount)
	svc.bot.Send(msg)

	return nil
}

func nsCount(update tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	count, err := GetNsCount()
	if err != nil {
		return fmt.Errorf("Can't select ns count from kube - %v\n", err)
	}
	msg.Text = fmt.Sprintf("Количество неймспейсов в кубе - %v\n", count)
	svc.bot.Send(msg)

	return nil
}

func deployCount(update tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	count, err := GetDeployCount()
	if err != nil {
		return fmt.Errorf("Can't select deployments count from kube - %v\n", err)
	}
	msg.Text = fmt.Sprintf("Количество деплойментов в кубе - %v\n", count)
	svc.bot.Send(msg)

	return nil
}

func getImages(update tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	active, notActive, err := GetImages()
	if err != nil {
		return fmt.Errorf("Can't select deployments from kube - %v\n", err)
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

func getEmailInfo(update tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

	mail := update.Message.CommandArguments()
	ns, err := GetUserInfo(mail)
	fmt.Println(ns)
	if err != nil {
		return fmt.Errorf("Can't get ns from email - %v\n", err)
	}
	deployments, err := GetUserDeploys(ns)
	if err != nil {
		return fmt.Errorf("Can't get deployments from ns - %v\n", err)
	}
	fmt.Println(deployments)
	s := ""
	for name, ds := range deployments {
		s += fmt.Sprintf("\n\nNamespace - %s\n\n", name)
		for _, deploy := range ds {
			if deploy.IsActive {
				s += fmt.Sprintf("%02d - Running - %s\n", deploy.Replicas, deploy.Image)
			} else {
				s += fmt.Sprintf("%02d - Disable - %s\n", deploy.Replicas, deploy.Image)
			}
		}
	}
	msg.Text = s

	//msg.Text = fmt.Sprintf("Активные image: %v\n"+
	//	"Неактивные image:  %v\n", activeString, notActiveString)
	svc.bot.Send(msg)

	return nil
}
