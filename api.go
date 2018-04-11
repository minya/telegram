package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/minya/goutils/web"
	"net/http"
)

func SendMessage(botToken string, msg ReplyMessage) error {
	url := getMethodUrl(botToken, "sendMessage")
	client := http.Client{
		Transport: web.DefaultTransport(1000),
	}
	messageBytes, err := json.Marshal(msg)
	response, err := client.Post(url, "application/json", bytes.NewReader(messageBytes))
	if err != nil {
		return err
	}
	if response.StatusCode >= 400 {
		return fmt.Errorf("%v from telegram API", response.StatusCode)
	}
	return nil
}

func getMethodUrl(botToken string, methodName string) string {
	return fmt.Sprintf("https://api.telegram.org/bot%v/%v", botToken, methodName)

}
