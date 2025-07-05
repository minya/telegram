package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/minya/goutils/web"
	"github.com/minya/logger"
)

const (
	telegramAPIErrorFmt = "telegram api error: %v"
)

var updatesHandler func(Update) any
var apiToken string

// StartListen runs handler
func StartListen(botAPIToken string, port int, handler func(Update) interface{}) error {
	updatesHandler = handler
	apiToken = botAPIToken
	http.HandleFunc("/", handleHTTP)
	logger.Info(fmt.Sprintf("Listen on %v\n", port))
	return http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
}

func handleHTTP(w http.ResponseWriter, r *http.Request) {
	bytes, _ := io.ReadAll(r.Body)
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

func sendMessage(msg ReplyMessage) error {
	client := http.Client{
		Transport: web.DefaultTransport(1000),
	}

	sendMsgURL := fmt.Sprintf("https://api.telegram.org/bot%v/sendMessage", apiToken)

	logger.Info(fmt.Sprintf("Sending msg to %v\n", msg.ChatId))
	msgBin, _ := json.Marshal(msg)
	bodyReader := bytes.NewReader(msgBin)
	resp, err := client.Post(sendMsgURL, "application/json", bodyReader)
	logger.Info(fmt.Sprintf("%v from telegram api\n", resp.StatusCode))
	if nil != err {
		logger.Error(err, "Send message failed")
		return fmt.Errorf("send message failed: %v", err)
	}

	if resp.StatusCode >= 400 {
		logger.Error(fmt.Errorf(telegramAPIErrorFmt, resp.StatusCode), "Send message failed")
		return fmt.Errorf(telegramAPIErrorFmt, resp.StatusCode)
	}
	return nil
}

func sendDocument(document ReplyDocument) error {
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

	resp, err := client.Post(sendMsgURL, mpWriter.FormDataContentType(), &buf)
	if err != nil {
		logger.Error(err, "Send document failed")
		return fmt.Errorf("send document failed: %v", err)
	}
	if resp.StatusCode >= 400 {
		logger.Error(fmt.Errorf(telegramAPIErrorFmt, resp.StatusCode), "Send document failed")
		return fmt.Errorf(telegramAPIErrorFmt, resp.StatusCode)
	}
	return nil
}
