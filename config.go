package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Config struct {
	TelegramBot TelegramConfig
	Clickhouse  ClickhouseConfig
	Kube        KubeConfig
	Postgres    PostgresConfig
}

type KubeConfig struct {
	Path string
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

func OpenConfig(path string) Config {
	var config Config
	configFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read %s", err)
	}
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatalf("Failed to parse %s", err)
	}
	return config
}
