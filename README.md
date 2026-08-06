# Raindrop Telegram Bot

A Telegram bot that saves links to [Raindrop.io](https://raindrop.io). Send it any URL and it lands in your Unsorted collection.

Built as a learning project after Mozilla shut down Pocket.

## Features

- OAuth2 authorization — each user connects their own Raindrop account
- Tokens persisted in an embedded [bbolt](https://github.com/etcd-io/bbolt) database, survive restarts
- Automatic token refresh — a two-week access token is renewed transparently
- Bookmarks are created with `pleaseParse`, so Raindrop fetches the title, description and cover itself
- Single binary, no external services

## How it works

```
1. user: /start                 →  bot
2. bot: authorization link      →  user          (nothing is sent anywhere yet)
3. user clicks                  →  browser opens raindrop.io
4. user grants access           →  Raindrop redirects back
5. browser                      →  bot's HTTP server /callback?code=...&state=...
6. bot server ── POST ──────────→  raindrop.io/oauth/access_token
7. bot stores tokens in bbolt, keyed by chat_id
```

The `state` parameter carries a random 32-byte value generated with `crypto/rand`. The bot keeps a one-time `state → chat_id` mapping in bbolt and deletes it on first use, so a callback cannot be forged for another user's chat.

## Requirements

- Go 1.26+
- A Telegram bot token from [@BotFather](https://t.me/BotFather)
- A Raindrop app from [app.raindrop.io/settings/integrations](https://app.raindrop.io/settings/integrations)

## Setup

Create the Raindrop app and set its **Redirect URI** to exactly:

```
http://localhost:8080/callback
```

Raindrop matches this string exactly — any extra query parameter results in `Incorrect redirect_uri`.

Create a `.env` file in the project root:

```env
TELEGRAM_BOT_TOKEN=123456:ABC-DEF...
RAINDROP_CLIENT_ID=your_client_id
RAINDROP_CLIENT_SECRET=your_client_secret
RAINDROP_REDIRECT_URI=http://localhost:8080/callback
RAINDROP_TEST_TOKEN=your_test_token
```

`.env` is gitignored. Never commit it.

## Running

```bash
make run      # build + run
make build    # build only, binary lands in ./.bin/bot
```

The bot starts a long-polling loop for Telegram and an HTTP server on `:8080` for the OAuth callback. Both run in the same process.

## Bot commands

| Command | Description |
|---|---|
| `/start` | Authorize in Raindrop |
| `/status` | Check authorization and expiry date |
| `/logout` | Revoke access, delete stored tokens |

Any message containing an `http://` or `https://` link is saved as a bookmark.

## Project layout

```
cmd/bot/          entry point, wires everything together
pkg/config/       environment variables
pkg/telegram/     bot: update loop, commands, message handling
pkg/server/       HTTP server for the OAuth callback
pkg/raindrop/     Raindrop API client (Authorize, Refresh, AddBookmark)
pkg/storage/      bbolt storage for tokens and OAuth states
```

## Notes

The database file `bot.db` is created next to the binary on first run and is gitignored. It is locked by a single process — a second instance will hang waiting for the lock.

Raindrop does not document the `state` parameter, but it does pass it through the authorization flow.

## Deployment

For a public deployment, change `RAINDROP_REDIRECT_URI` to a reachable HTTPS address and update the Redirect URI in the Raindrop app settings to match it exactly.
