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
	_, err := api.callMethod("sendMessage", msg)
	if err != nil {
		return err
	}
	return nil
}

func (api *Api) GetUpdates(offset int) ([]Update, error) {
	type getUpdatesParams struct {
		Offset  int    `json:"offset"`
		Timeout uint32 `json:"timeout"`
	}
	msg := getUpdatesParams{
		Offset:  offset,
		Timeout: 1,
	}
	responseBytes, err := api.callMethod("getUpdates", msg)
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

func (api *Api) GetFile(fileID string) (File, error) {
	type getFileArgs struct {
		FileID string `json:"file_id"`
	}
	type getFileResponse struct {
		Ok     bool `json:"ok"`
		Result File `json:"result"`
	}
	responseBytes, err := api.callMethod("getFile", getFileArgs{FileID: fileID})
	var response getFileResponse
	if err != nil {
		return response.Result, err
	}
	err = json.Unmarshal(responseBytes, &response)
	return response.Result, err
}

func (api *Api) DownloadFile(file File) ([]byte, error) {
	url := fmt.Sprintf("https://api.telegram.org/file/bot%v/%v", api.botToken, *file.FilePath)
	resp, err := api.client.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%v from api", resp.StatusCode)
	}
	return ioutil.ReadAll(resp.Body)
}

func getMethodUrl(botToken string, methodName string) string {
	return fmt.Sprintf("https://api.telegram.org/bot%v/%v", botToken, methodName)
}

func (api *Api) callMethod(methodName string, payload interface{}) ([]byte, error) {
	url := getMethodUrl(api.botToken, methodName)
	messageBytes, err := json.Marshal(payload)
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
	return responseBytes, err
}

type UpdatesResult struct {
	Ok     bool     `json:"ok"`
	Result []Update `json:"result,omitempty"`
}
