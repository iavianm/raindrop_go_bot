package telegram

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/url"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const helpText = "Here is what I can do:\n\n" +
	"/start - authorize in Raindrop\n" +
	"/status - check your authorization\n" +
	"/logout - revoke access\n" +
	"/help - show this message\n\n" +
	"Send me any link and I will save it to Raindrop."

func (b *Bot) handleUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		if update.Message != nil {
			if update.Message.IsCommand() {
				err := b.handleCommand(update.Message)
				if err != nil {
					log.Println(err)
				}
				continue
			}

			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			if err := b.handleMessage(update.Message); err != nil {
				log.Println(err)
			}
		}
	}
}

func (b *Bot) handleMessage(message *tgbotapi.Message) error {
	tokens, err := b.validTokens(message.Chat.ID)
	if err != nil {
		b.reply(message.Chat.ID, "Couldn't refresh your access. Send /start to authorize again.")
		return err
	}
	if tokens == nil {
		b.reply(message.Chat.ID, "You are not authorized yet. Send /start first.")
		return nil
	}

	u, err := url.ParseRequestURI(message.Text)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		b.reply(message.Chat.ID, "Please send me a link starting with http:// or https://")
		return nil
	}

	err = b.raindrop.AddBookmark(tokens.AccessToken, u.String())
	if err != nil {
		b.reply(message.Chat.ID, "Couldn't save the link. Please try again later.")
		return err
	}

	b.reply(message.Chat.ID, "Saved to Raindrop.")
	return nil
}

func (b *Bot) reply(chatID int64, text string) {
	if _, err := b.bot.Send(tgbotapi.NewMessage(chatID, text)); err != nil {
		log.Println(err)
	}
}

func (b *Bot) handleCommand(message *tgbotapi.Message) error {
	switch message.Command() {
	case "start":
		query := url.Values{}
		query.Set("client_id", b.env.RaindropClientID)
		query.Set("redirect_uri", b.env.RaindropRedirectURI)
		query.Set("response_type", "code")
		state, err := generateState()
		if err != nil {
			return err
		}
		err = b.storage.SaveState(state, message.Chat.ID)
		if err != nil {
			return err
		}
		query.Set("state", state)

		u := &url.URL{
			Scheme:   "https",
			Host:     "raindrop.io",
			Path:     "oauth/authorize",
			RawQuery: query.Encode(),
		}

		b.reply(message.Chat.ID, u.String())
	case "status":
		tokens, err := b.validTokens(message.Chat.ID)
		if err != nil {
			b.reply(message.Chat.ID, "Couldn't refresh your access. Send /start to authorize again.")
			return err
		}
		if tokens == nil {
			b.reply(message.Chat.ID, "You are not authorized yet. Send /start first.")
			return nil
		}

		b.reply(message.Chat.ID, fmt.Sprintf("You are authorized. Access is valid until %s.",
			tokens.ExpiresAt.Format(time.RFC1123)))
	case "logout":
		err := b.storage.DeleteTokens(message.Chat.ID)
		if err != nil {
			b.reply(message.Chat.ID, "Couldn't log you out. Please try again later.")
			return err
		}
		b.reply(message.Chat.ID, "You have been logged out.")
	case "help":
		b.reply(message.Chat.ID, helpText)
	default:
		b.reply(message.Chat.ID, "Unknown command.\n\n"+helpText)
	}

	return nil
}

func generateState() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
