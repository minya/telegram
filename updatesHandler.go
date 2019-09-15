package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/minya/goutils/web"
)

var updatesHandler func(Update) interface{}
var apiToken string

// StartListen runs handler
func StartListen(botAPIToken string, port int, handler func(Update) interface{}) error {
	updatesHandler = handler
	apiToken = botAPIToken
	http.HandleFunc("/", handleHTTP)
	log.Printf("Listen on %v\n", port)
	return http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
}

func handleHTTP(w http.ResponseWriter, r *http.Request) {
	bytes, _ := ioutil.ReadAll(r.Body)
	var upd Update
	json.Unmarshal(bytes, &upd)

	var response interface{} = updatesHandler(upd)

	message, ok := response.(ReplyMessage)
	if ok {
		sendMessage(message)
		io.WriteString(w, "ok")
		return
	}
	document, ok := response.(ReplyDocument)
	if ok {
		sendDocument(document)
		io.WriteString(w, "ok")
		return
	}
}

func sendMessage(msg ReplyMessage) {
	client := http.Client{
		Transport: web.DefaultTransport(1000),
	}

	sendMsgURL := fmt.Sprintf("https://api.telegram.org/bot%v/sendMessage", apiToken)

	log.Printf("Sending msg to %v\n", msg.ChatId)
	msgBin, _ := json.Marshal(msg)
	bodyReader := bytes.NewReader(msgBin)
	resp, err := client.Post(sendMsgURL, "application/json", bodyReader)
	if nil != err {
		log.Printf("%v\n", err)
		return
	}

	log.Printf("%v from telegram api\n", resp.StatusCode)
}

func sendDocument(document ReplyDocument) {
	client := http.Client{
		Transport: web.DefaultTransport(1000),
	}

	sendMsgURL := fmt.Sprintf("https://api.telegram.org/bot%v/sendDocument", apiToken)

	var buf bytes.Buffer
	mpWriter := multipart.NewWriter(&buf)
	fw, _ := mpWriter.CreateFormFile("document", document.InputFile.FileName)
	fw.Write(document.InputFile.Content)
	mpWriter.WriteField("chat_id", strconv.Itoa(document.ChatId))
	mpWriter.WriteField("caption", document.Caption)
	mpWriter.Close()

	//response, _ :=
	response, _ := client.Post(sendMsgURL, mpWriter.FormDataContentType(), &buf)
	rb, _ := ioutil.ReadAll(response.Body)
	fmt.Printf("%v\n%v\n", response, string(rb))
}
