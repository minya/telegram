package telegram

import (
	"fmt"
	"time"

	"github.com/minya/logger"
)

type UpdateHandler func(upd *Update) error

func StartPolling(api *Api, handle UpdateHandler, updateInterval time.Duration, offset int) error {
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
				api.SendMessage(ReplyMessage{
					ChatId: upd.Message.From.Id,
					Text:   "Error while handling update",
				})
			}

			newOffset = upd.UpdateId + 1
		}

		time.Sleep(updateInterval)
	}
}
