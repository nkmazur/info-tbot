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
}

type KubeConfig struct {
	KubeConfig string
}

type TelegramConfig struct {
	Token string
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

func OpenMessage(path string) string {
	message, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read %s", err)
	}
	return string(message)
}
