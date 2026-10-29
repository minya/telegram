# telegram

A small Go client for the [Telegram Bot API](https://core.telegram.org/bots/api).
It covers the methods needed for typical bots: receiving updates by long
polling or webhook, sending messages, documents and photos, editing and
deleting messages, downloading files, keyboards, callback queries, commands and
Web App menu buttons.

```
go get github.com/minya/telegram
```

Requires Go 1.23+.

## Quick start

```go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/minya/telegram"
)

func main() {
	api := telegram.NewAPI(os.Getenv("TELEGRAM_BOT_TOKEN"))

	handler := func(ctx context.Context, upd *telegram.Update) error {
		if upd.Message == nil {
			return nil
		}
		_, err := api.SendMessage(ctx, telegram.SendMessageParams{
			ChatID: upd.Message.Chat.ID,
			Text:   "You said: " + upd.Message.Text,
		})
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	err := telegram.Poll(ctx, api, handler, telegram.PollOptions{
		OnError: func(upd *telegram.Update, err error) { log.Print(err) },
	})
	log.Print(err) // context.Canceled after Ctrl+C
}
```

## Client

```go
api := telegram.NewAPI(token)
api := telegram.NewAPI(token,
	telegram.WithBaseURL("http://localhost:8081"), // local Bot API server or test server
	telegram.WithHTTPClient(myClient),
)
```

The default HTTP client has a 60 second timeout. Every method takes a
`context.Context` for cancellation and deadlines.

| Method | Returns |
| --- | --- |
| `GetMe` | `*User` |
| `GetUpdates(GetUpdatesParams)` | `[]Update` |
| `SendMessage(SendMessageParams)` | `*Message` |
| `SendDocument(SendDocumentParams)` | `*Message` |
| `SendPhoto(SendPhotoParams)` | `*Message` |
| `ForwardMessage(ForwardMessageParams)` | `*Message` |
| `CopyMessage(CopyMessageParams)` | new message ID |
| `EditMessageText(EditMessageTextParams)` | `*Message` (nil for inline messages) |
| `EditMessageReplyMarkup(EditMessageReplyMarkupParams)` | `*Message` (nil for inline messages) |
| `DeleteMessage(chatID, messageID)` | |
| `SendChatAction(chatID, action)` | |
| `AnswerCallbackQuery(AnswerCallbackQueryParams)` | |
| `SetMyCommands(SetMyCommandsParams)` | |
| `SetChatMenuButton(SetChatMenuButtonParams)` | |
| `SetWebhook(SetWebhookParams)` / `DeleteWebhook` / `GetWebhookInfo` | |
| `GetFile(fileID)` | `*File` |
| `DownloadFile(*File)` | `[]byte` |

### Errors

When Telegram rejects a request, the method returns a `*telegram.Error` with
the method name, error code, description and, for flood control, `RetryAfter`:

```go
var apiErr *telegram.Error
if errors.As(err, &apiErr) && apiErr.Code == 429 {
	time.Sleep(time.Duration(apiErr.RetryAfter) * time.Second)
}
```

The bot token is removed from all returned errors, so they are safe to log.

## Receiving updates

### Long polling

```go
func Poll(ctx context.Context, api *API, handle UpdateHandler, opts PollOptions) error
```

`Poll` long-polls `getUpdates` (50 second timeout by default) and calls
`handle` for each update until `ctx` is cancelled. If `getUpdates` fails, it
retries with exponential backoff from 1 to 30 seconds, and waits for
`retry_after` when Telegram asks it to. Both polling errors and handler errors
go to `PollOptions.OnError`. For polling errors the `upd` argument is nil.

### Webhook

```go
api.SetWebhook(ctx, telegram.SetWebhookParams{
	URL:         "https://example.com/telegram/webhook",
	SecretToken: "some-secret",
})

http.Handle("/telegram/webhook", telegram.WebhookHandler("some-secret", handler, onError))
```

`WebhookHandler` accepts only POST requests that carry the right
`X-Telegram-Bot-Api-Secret-Token` header. It decodes the update and calls the
handler. It answers 200 even when the handler fails, so Telegram does not keep
redelivering the update. To upload a self-signed certificate, set
`SetWebhookParams.Certificate`.

## Examples

### Inline keyboard and callback queries

```go
api.SendMessage(ctx, telegram.SendMessageParams{
	ChatID: chatID,
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
	api.AnswerCallbackQuery(ctx, telegram.AnswerCallbackQueryParams{CallbackQueryID: q.ID})
	api.EditMessageText(ctx, telegram.EditMessageTextParams{
		ChatID:    q.Message.Chat.ID,
		MessageID: q.Message.MessageID,
		Text:      "You picked: " + q.Data,
	})
}
```

`ReplyMarkup` accepts `InlineKeyboardMarkup`, `ReplyKeyboardMarkup`,
`ReplyKeyboardRemove` or `ForceReply`.

### Receiving and sending files

`InputFile` either uploads new data (`FileName` + `Content`) or refers to an
existing file by `file_id` or URL (`Ref`).

```go
if upd.Message.HasDocument() {
	file, err := api.GetFile(ctx, upd.Message.Document.FileID)
	if err != nil {
		return err
	}
	content, err := api.DownloadFile(ctx, file)
	if err != nil {
		return err
	}

	_, err = api.SendDocument(ctx, telegram.SendDocumentParams{
		ChatID:   upd.Message.Chat.ID,
		Caption:  "Here it is back",
		Document: telegram.InputFile{FileName: upd.Message.Document.FileName, Content: content},
	})
	return err
}

// Re-send a photo without uploading it again:
api.SendPhoto(ctx, telegram.SendPhotoParams{
	ChatID: chatID,
	Photo:  telegram.InputFile{Ref: upd.Message.Photo[0].FileID},
})
```

Telegram does not allow bots to download files larger than 20 MB. For those,
`File.FilePath` is empty and `DownloadFile` returns an error.

### Web App menu button

```go
api.SetChatMenuButton(ctx, telegram.SetChatMenuButtonParams{
	MenuButton: &telegram.MenuButton{
		Type:   "web_app",
		Text:   "Open app",
		WebApp: &telegram.WebAppInfo{URL: "https://example.com/app"},
	},
})
```

Leave `ChatID` as zero to change the bot's default menu button. Inline buttons
can also open a Web App via `InlineKeyboardButton.WebApp`.

## Types

`types.go` defines the subset of Bot API objects this client uses, with fields
named after the Bot API's JSON fields. It also has constants for parse modes
(`ParseModeHTML`, `ParseModeMarkdownV2`), chat types (`ChatTypePrivate`, …)
and chat actions (`ChatActionTyping`, …).

## Testing

```
go test -race ./...
```

Tests run against an `httptest` server through `WithBaseURL`. They make no
real Telegram requests.
