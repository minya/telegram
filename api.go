package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/minya/goutils/web"
)

type Api struct {
	botToken string
	client   http.Client
}

func NewApi(botToken string) Api {
	client := http.Client{
		Transport: web.DefaultTransport(1000),
	}
	return Api{
		botToken: botToken,
		client:   client,
	}
}

func (api *Api) SendMessage(msg ReplyMessage) error {
	url := getMethodUrl(api.botToken, "sendMessage")
	messageBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	response, err := api.client.Post(url, "application/json", bytes.NewReader(messageBytes))
	if err != nil {
		return err
	}
	if response.StatusCode >= 400 {
		return fmt.Errorf("%v from telegram API", response.StatusCode)
	}
	return nil
}

func (api *Api) GetUpdates(offset int) ([]Update, error) {
	url := getMethodUrl(api.botToken, "getUpdates")
	msg := getUpdatesParams{
		Offset:  offset,
		Timeout: 1,
	}
	messageBytes, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	response, err := api.client.Post(url, "application/json", bytes.NewReader(messageBytes))
	if err != nil {
		return nil, err
	}
	if response.StatusCode >= 400 {
		return nil, fmt.Errorf("%v from telegram API", response.StatusCode)
	}
	responseBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var responseObject UpdatesResult
	err = json.Unmarshal(responseBytes, &responseObject)
	if err != nil {
		return nil, err
	}
	if !responseObject.Ok {
		return nil, fmt.Errorf("result wasn't ok")
	}

	return responseObject.Result, nil
}

func getMethodUrl(botToken string, methodName string) string {
	return fmt.Sprintf("https://api.telegram.org/bot%v/%v", botToken, methodName)
}

type getUpdatesParams struct {
	Offset  int    `json:"offset"`
	Timeout uint32 `json:"timeout"`
}

type UpdatesResult struct {
	Ok     bool     `json:"ok"`
	Result []Update `json:"result,omitempty"`
}
