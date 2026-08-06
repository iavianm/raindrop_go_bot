package config

import (
	"log"
	"os"
)

type ENV struct {
	TelegramBotToken     string
	RaindropClientID     string
	RaindropClientSecret string
	RaindropRedirectURI  string
	RaindropTestToken    string
}

func NewEnv() *ENV {
	env := new(ENV)
	env.TelegramBotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	if env.TelegramBotToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable not set")
	}

	env.RaindropClientID = os.Getenv("RAINDROP_CLIENT_ID")
	if env.RaindropClientID == "" {
		log.Fatal("RAINDROP_CLIENT_ID must be set")
	}

	env.RaindropClientSecret = os.Getenv("RAINDROP_CLIENT_SECRET")
	if env.RaindropClientSecret == "" {
		log.Fatal("RAINDROP_CLIENT_SECRET must be set")
	}

	env.RaindropRedirectURI = os.Getenv("RAINDROP_REDIRECT_URI")
	if env.RaindropRedirectURI == "" {
		log.Fatal("RAINDROP_REDIRECT_URI must be set")
	}

	env.RaindropTestToken = os.Getenv("RAINDROP_TEST_TOKEN")
	if env.RaindropTestToken == "" {
		log.Fatal("RAINDROP_TEST_TOKEN environment variable not set")
	}
	return env
}
