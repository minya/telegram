package telegram

import (
	"log"
	"time"
)

type UpdateHandler func(upd *Update) error

func StartPolling(api *Api, handle UpdateHandler, updateInterval time.Duration, offset int) error {
	newOffset := offset
	for {
		//log.Printf("Getting updates with offset %d\n", newOffset)

		updates, err := api.GetUpdates(newOffset)
		if err != nil {
			log.Printf("Error while getting updates: %v\n", err)
			return err
		}
		for _, upd := range updates {
			err = handle(&upd)
			if err != nil {
				log.Printf("Error while handling update: %v\n", err)
				api.SendMessage(ReplyMessage{ // TODO: allow to disable this
					ChatId: upd.Message.From.Id,
					Text:   "Error while handling update",
				})
			}

			log.Printf("Update received %#v\n", upd)
			newOffset = upd.UpdateId + 1
		}

		//log.Printf("Updates handled, new offset is %d\n", newOffset)
		//log.Printf("Sleeping for %v\n", updateInterval)
		time.Sleep(updateInterval)
	}
}
