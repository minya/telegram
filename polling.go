package telegram

import (
	"log"
	"time"
)

type UpdateHandler func(upd *Update) interface{}

func StartPolling(api *Api, handle UpdateHandler, updateInterval time.Duration, offset int) error {
	newOffset := offset
	for {
		updates, err := api.GetUpdates(newOffset)
		if err != nil {
			return err
		}
		for _, upd := range updates {
			msg := handle(&upd)
			api.SendMessage(msg.(ReplyMessage))
			log.Printf("Update received %#v\n", upd)
			newOffset = upd.UpdateId + 1
		}
		time.Sleep(updateInterval)
	}
}
