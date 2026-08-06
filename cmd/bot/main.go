package main

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/iavianm/raindrop_go_bot/pkg/config"
	"github.com/iavianm/raindrop_go_bot/pkg/raindrop"
	"github.com/iavianm/raindrop_go_bot/pkg/server"
	"github.com/iavianm/raindrop_go_bot/pkg/storage"
	"github.com/iavianm/raindrop_go_bot/pkg/telegram"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	env := config.NewEnv()

	bot, err := tgbotapi.NewBotAPI(env.TelegramBotToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	db, err := storage.New("bot.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rd := raindrop.New(env.RaindropClientID, env.RaindropClientSecret, env.RaindropRedirectURI)

	go func() {
		err := telegram.NewBot(bot, env, rd, db).Start()
		if err != nil {
			log.Fatal(err)
		}
	}()

	server.NewServer(rd, db).Start()
}
