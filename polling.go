package telegram

import (
	"fmt"
	"time"

	"github.com/minya/logger"
)

type UpdateHandler func(upd *Update) error

func StartPolling(api *Api, handle UpdateHandler, updateInterval time.Duration, offset int64) error {
	newOffset := offset

	for {
		logger.Debug("Getting updates with offset %d", newOffset)

		updates, err := api.GetUpdates(newOffset)
		if err != nil {
			logger.Error(err, "Error while getting updates")
			return err
		}
			for _, upd := range updates {
				logger.Info(fmt.Sprintf("Update received %#v\n", upd))
				err = handle(&upd)
				if err != nil {
					logger.Error(err, "Error while handling update")
					var chatID int64
					switch {
					case upd.Message != nil:
						chatID = upd.Message.Chat.Id
					case upd.CallbackQuery != nil && upd.CallbackQuery.Message != nil:
						chatID = upd.CallbackQuery.Message.Chat.Id
					}
					if chatID != 0 {
						api.SendMessage(ReplyMessage{
							ChatId: chatID,
							Text:   "Error while handling update",
						})
					}
				}

				newOffset = upd.UpdateId + 1
			}

		time.Sleep(updateInterval)
	}
}
