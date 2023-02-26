package telegram

import (
	"log"
	"time"
)

type UpdateHandler func(upd *Update) error

func StartPolling(api *Api, handle UpdateHandler, updateInterval time.Duration, offset int, notify chan int) error {
	newOffset := offset
	for {
		updates, err := api.GetUpdates(newOffset)
		if err != nil {
			return err
		}
		for _, upd := range updates {
			err = handle(&upd)
			if err != nil {
				log.Printf("%#v\n", err)
				api.SendMessage(ReplyMessage{
					ChatId: upd.Message.From.Id,
					Text:   "При обработке запроса прооизошла ошибка",
				})
			}
			log.Printf("Update received %#v\n", upd)
			newOffset = upd.UpdateId + 1
		}
		if len(updates) > 0 {
			notify <- 1
		}
		time.Sleep(updateInterval)
	}
}
