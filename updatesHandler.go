package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/minya/goutils/web"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

var updatesHandler func(Update) ReplyMessage
var apiToken string

func StartListen(botId string, port int, handler func(Update) ReplyMessage) error {
	updatesHandler = handler
	apiToken = botId
	http.HandleFunc("/", handleHttp)
	log.Printf("Listen on %v\n", port)
	return http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
}

func handleHttp(w http.ResponseWriter, r *http.Request) {
	bytes, _ := ioutil.ReadAll(r.Body)
	var upd Update
	json.Unmarshal(bytes, &upd)

	message := updatesHandler(upd)

	sendMessage(upd.Message.Chat.Id, message)

	io.WriteString(w, "ok")

}

func sendMessage(chatId int, msg ReplyMessage) {
	client := http.Client{
		Transport: web.DefaultTransport(1000),
	}

	sendMsgUrl := fmt.Sprintf("https://api.telegram.org/bot%v/sendMessage", apiToken)

	log.Printf("Sending msg to %v\n", chatId)
	msgBin, _ := json.Marshal(msg)
	bodyReader := bytes.NewReader(msgBin)
	resp, err := client.Post(sendMsgUrl, "application/json", bodyReader)
	if nil != err {
		log.Printf("%v\n", err)
		return
	}

	log.Printf("%v from telegram api\n", resp.StatusCode)
}
