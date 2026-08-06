package telegram

import (
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/iavianm/raindrop_go_bot/pkg/config"
	"github.com/iavianm/raindrop_go_bot/pkg/raindrop"
	"github.com/iavianm/raindrop_go_bot/pkg/storage"
)

type Bot struct {
	bot      *tgbotapi.BotAPI
	env      *config.ENV
	storage  *storage.Storage
	raindrop *raindrop.Client
}

func NewBot(bot *tgbotapi.BotAPI, env *config.ENV, rd *raindrop.Client, db *storage.Storage) *Bot {
	return &Bot{
		bot:      bot,
		env:      env,
		storage:  db,
		raindrop: rd,
	}
}

func (b *Bot) Start() error {
	log.Printf("Authorized on account %s", b.bot.Self.UserName)

	cmds := tgbotapi.NewSetMyCommands(
		tgbotapi.BotCommand{Command: "start", Description: "Authorize in Raindrop"},
		tgbotapi.BotCommand{Command: "status", Description: "Check your authorization"},
		tgbotapi.BotCommand{Command: "logout", Description: "Revoke access"},
	)
	if _, err := b.bot.Request(cmds); err != nil {
		return err
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.bot.GetUpdatesChan(u)

	b.handleUpdates(updates)

	return nil
}

func (b *Bot) validTokens(chatID int64) (*storage.StoredTokens, error) {
	tokens, err := b.storage.GetTokens(chatID)
	if err != nil {
		return nil, err
	}

	if tokens == nil {
		return nil, nil
	}

	expired := time.Now().After(tokens.ExpiresAt.Add(-5 * time.Minute))
	if expired {
		newTokens, err := b.raindrop.Refresh(tokens.RefreshToken)
		if err != nil {
			return nil, fmt.Errorf("refresh tokens: %w", err)
		}

		if newTokens.RefreshToken == "" {
			newTokens.RefreshToken = tokens.RefreshToken
		}

		err = b.storage.SaveTokens(chatID, newTokens)
		if err != nil {
			return nil, err
		}

		tokens, err := b.storage.GetTokens(chatID)
		if err != nil {
			return nil, err
		}

		if tokens == nil {
			return nil, nil
		}

		return tokens, nil
	}

	return tokens, nil
}
