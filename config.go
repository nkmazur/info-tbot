package main

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	TelegramBot TelegramConfig
	Clickhouse  ClickhouseConfig
	Postgres    PostgresConfig
}

type PostgresConfig struct {
	Url string
}

type TelegramConfig struct {
	Token string
	Users map[string]string
}

type ClickhouseConfig struct {
	Url string
}

func OpenConfig() Config {
	var config Config
	config.Postgres.Url = fmt.Sprintf("host=%s user=%s password=%s dbname=%s",
		os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASS"), os.Getenv("POSTGRES_DB"))
	config.Clickhouse.Url = fmt.Sprintf("tcp://%s?username=%s&password=%s&database=%s&debug=true",
		os.Getenv("CLICKHOUSE_HOST"), os.Getenv("CLICKHOUSE_USER"), os.Getenv("CLICKHOUSE_PASS"), os.Getenv("CLICKHOUSE_DB"))
	config.TelegramBot.Token = os.Getenv("TELEGRAM_TOKEN")
	config.TelegramBot.Users = ParseUsers()
	return config
}
func ParseUsers() map[string]string {
	usersStr := os.Getenv("ALLOW_USERS")
	var users map[string]string
	users = make(map[string]string, 1)
	for _, userPass := range strings.Split(usersStr, ",") {
		userBundle := strings.Split(userPass, ":")
		users[userBundle[1]] = userBundle[0]
	}
	return users
}
