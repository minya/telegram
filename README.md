# telegram

A small Go client for the [Telegram Bot API](https://core.telegram.org/bots/api).
It covers the methods needed for simple bots: receiving updates by long polling,
sending messages and documents, downloading files, inline keyboards, callback
queries, webhooks and Web App menu buttons.

```
go get github.com/minya/telegram
```

Requires Go 1.23+.

## Quick start

```go
package main

import (
	"log"
	"os"
	"time"

	"github.com/minya/telegram"
)

func main() {
	api := telegram.NewApi(os.Getenv("TELEGRAM_BOT_TOKEN"))

	handler := func(upd *telegram.Update) error {
		if upd.Message == nil {
			return nil
		}
		return api.SendMessage(telegram.ReplyMessage{
			ChatId: upd.Message.Chat.Id,
			Text:   "You said: " + upd.Message.Text,
		})
	}

	// Poll every second, starting from offset 0.
	if err := telegram.StartPolling(&api, handler, time.Second, 0); err != nil {
		log.Fatal(err)
	}
}
```

## API

Create a client with `telegram.NewApi(botToken)`. All methods are on `*Api`.

| Method | Telegram method | Notes |
| --- | --- | --- |
| `SendMessage(ReplyMessage) error` | `sendMessage` | |
| `GetUpdates(offset int64) ([]Update, error)` | `getUpdates` | Uses a 1 second long-poll timeout. |
| `GetFile(fileID string) (File, error)` | `getFile` | |
| `DownloadFile(File) ([]byte, error)` | — | Fetches `File.FilePath` from the file endpoint. |
| `SendDocument(ReplyDocument) error` | `sendDocument` | Uploads `InputFile` as multipart form data. |
| `AnswerCallbackQuery(*AnswerCallbackQueryParams) error` | `answerCallbackQuery` | |
| `EditMessageText(*EditMessageTextParams) error` | `editMessageText` | |
| `SetWebhook(*SetWebhookParams) error` | `setWebhook` | |
| `SetChatMenuButton(*SetChatMenuButtonParams) error` | `setChatMenuButton` | |

### Polling

```go
func StartPolling(api *Api, handle UpdateHandler, updateInterval time.Duration, offset int64) error
```

`StartPolling` loops forever: it fetches updates, calls `handle` for each one,
advances the offset, then sleeps for `updateInterval`.

- If `handle` returns an error, the error is logged and the bot replies
  "Error while handling update" to the originating chat (for messages and
  callback queries). Polling continues.
- If `getUpdates` itself fails, `StartPolling` returns that error.

## Examples

### Inline keyboard and callback queries

```go
api.SendMessage(telegram.ReplyMessage{
	ChatId: chatID,
	Text:   "Pick one",
	ReplyMarkup: telegram.InlineKeyboardMarkup{
		InlineKeyboard: [][]telegram.InlineKeyboardButton{{
			{Text: "Yes", CallbackData: "yes"},
			{Text: "No", CallbackData: "no"},
		}},
	},
})

// In the update handler:
if q := upd.CallbackQuery; q != nil {
	api.AnswerCallbackQuery(&telegram.AnswerCallbackQueryParams{CallbackQueryID: q.Id})
	api.EditMessageText(&telegram.EditMessageTextParams{
		ChatID:    q.Message.Chat.Id,
		MessageID: q.Message.MessageId,
		Text:      "You picked: " + q.Data,
	})
}
```

`ReplyMarkup` accepts either `InlineKeyboardMarkup` or `ReplyKeyboardMarkup`.

### Receiving and sending documents

```go
if upd.Message.HasDocument() {
	file, err := api.GetFile(upd.Message.Document.FileID)
	if err != nil {
		return err
	}
	content, err := api.DownloadFile(file)
	if err != nil {
		return err
	}

	return api.SendDocument(telegram.ReplyDocument{
		ChatId:  upd.Message.Chat.Id,
		Caption: "Here it is back",
		InputFile: telegram.InputFile{
			FileName: upd.Message.Document.FileName,
			Content:  content,
		},
	})
}
```

### Web App menu button

```go
api.SetChatMenuButton(&telegram.SetChatMenuButtonParams{
	MenuButton: &telegram.MenuButton{
		Type:   "web_app",
		Text:   "Open app",
		WebApp: &telegram.WebAppInfo{Url: "https://example.com/app"},
	},
})
```

Leave `ChatID` as zero to change the bot's default menu button. Inline buttons
can also open a Web App via `InlineKeyboardButton.WebApp`.

### Webhook

```go
api.SetWebhook(&telegram.SetWebhookParams{
	Url:         "https://example.com/telegram/webhook",
	SecretToken: "some-secret",
})
```

The library only registers the webhook; serving the endpoint and decoding the
incoming JSON into a `telegram.Update` is up to you.

## Types

`types.go` defines the subset of Bot API objects the client uses: `Update`,
`Message`, `User`, `Chat`, `MessageEntity`, `Document`, `PhotoSize`, `File`,
`CallbackQuery`, keyboard and button types, and the parameter structs for each
method. Fields follow the Bot API's JSON names.

## Dependencies

- [`github.com/minya/goutils`](https://github.com/minya/goutils) — HTTP transport
- [`github.com/minya/logger`](https://github.com/minya/logger) — logging
