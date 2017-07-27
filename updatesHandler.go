package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/minya/goutils/web"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
)

var updatesHandler func(Update) interface{}
var apiToken string

func StartListen(botId string, port int, handler func(Update) interface{}) error {
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

	var response interface{} = updatesHandler(upd)

	message, ok := response.(ReplyMessage)
	if ok {
		sendMessage(upd.Message.Chat.Id, message)
		io.WriteString(w, "ok")
		return
	}
	document, ok := response.(ReplyDocument)
	if ok {
		sendDocument(upd.Message.Chat.Id, document)
		io.WriteString(w, "ok")
		return
	}
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

func sendDocument(chatId int, document ReplyDocument) {
	client := http.Client{
		Transport: web.DefaultTransport(1000),
	}

	sendMsgUrl := fmt.Sprintf("https://api.telegram.org/bot%v/sendDocument", apiToken)

	var buf bytes.Buffer
	mpWriter := multipart.NewWriter(&buf)
	fw, _ := mpWriter.CreateFormFile("document", document.InputFile.FileName)
	fw.Write(document.InputFile.Content)
	mpWriter.WriteField("chat_id", strconv.Itoa(chatId))
	mpWriter.WriteField("caption", document.Caption)
	mpWriter.Close()

	//response, _ :=
	client.Post(sendMsgUrl, mpWriter.FormDataContentType(), &buf)
	//rb, _ := ioutil.ReadAll(response.Body)
	//fmt.Printf("%v\n%v\n", response, string(rb))
}
