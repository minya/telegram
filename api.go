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
	apiResponse, err := api.callMethod("sendMessage", msg)
	if err != nil {
		return err
	}
	if !apiResponse.Ok {
		return formatApiResponseError(apiResponse)
	}
	return nil
}

func (api *Api) GetUpdates(offset int64) ([]Update, error) {
	type getUpdatesParams struct {
		Offset  int64  `json:"offset"`
		Timeout uint32 `json:"timeout"`
	}
	msg := getUpdatesParams{
		Offset:  offset,
		Timeout: 1,
	}
	apiResponse, err := api.callMethod("getUpdates", msg)
	if err != nil {
		return nil, err
	}
	if !apiResponse.Ok {
		return nil, formatApiResponseError(apiResponse)
	}
	var updates []Update
	err = json.Unmarshal(apiResponse.Result, &updates)
	if err != nil {
		return nil, fmt.Errorf("Can't unmarshal response payload from telegram API: %v", err)
	}
	return updates, nil
}

func (api *Api) GetFile(fileID string) (File, error) {
	type getFileArgs struct {
		FileID string `json:"file_id"`
	}
	var file = File{}
	apiResponse, err := api.callMethod("getFile", getFileArgs{FileID: fileID})
	if err != nil {
		return file, err
	}
	if !apiResponse.Ok {
		return file, formatApiResponseError(apiResponse)
	}
	err = json.Unmarshal(apiResponse.Result, &file)
	if err != nil {
		return file, fmt.Errorf("Can't unmarshal response payload from telegram API: %v", err)
	}

	return file, err
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
	return io.ReadAll(resp.Body)
}

func (api *Api) AnswerCallbackQuery(params *AnswerCallbackQueryParams) error {
	_, err := api.callMethod("answerCallbackQuery", params)
	return err
}

func (api *Api) EditMessageText(params *EditMessageTextParams) error {
	_, err := api.callMethod("editMessageText", params)
	return err
}

func (api *Api) SendDocument(document ReplyDocument) error {
	client := http.Client{
		Transport: web.DefaultTransport(1000),
	}

	sendMsgURL := fmt.Sprintf("https://api.telegram.org/bot%v/sendDocument", api.botToken)

	var buf bytes.Buffer
	mpWriter := multipart.NewWriter(&buf)
	fw, _ := mpWriter.CreateFormFile("document", document.InputFile.FileName)
	fw.Write(document.InputFile.Content)
	mpWriter.WriteField("chat_id", strconv.FormatInt(document.ChatId, 10))
	mpWriter.WriteField("caption", document.Caption)
	if document.ParseMode != "" {
		mpWriter.WriteField("parse_mode", document.ParseMode)
	}
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

func (api *Api) SetWebhook(params *SetWebhookParams) error {
	_, err := api.callMethod("setWebhook", params)
	return err
}

func (api *Api) SetChatMenuButton(params *SetChatMenuButtonParams) error {
	_, err := api.callMethod("setChatMenuButton", params)
	return err
}

func getMethodUrl(botToken string, methodName string) string {
	return fmt.Sprintf("https://api.telegram.org/bot%v/%v", botToken, methodName)
}

func (api *Api) callMethod(methodName string, payload any) (ApiResponse, error) {
	url := getMethodUrl(api.botToken, methodName)
	messageBytes, err := json.Marshal(payload)
	var apiResponse = ApiResponse{}
	if err != nil {
		return apiResponse, err
	}
	response, err := api.client.Post(url, "application/json", bytes.NewReader(messageBytes))
	if err != nil {
		return apiResponse, err
	}
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return apiResponse, fmt.Errorf("Can't read response	from telegram API: %v", err)
	}
	err = json.Unmarshal(responseBytes, &apiResponse)
	if err != nil {
		return apiResponse, fmt.Errorf("Can't unmarshal response from telegram API: %v", err)
	}
	return apiResponse, err
}

func formatApiResponseError(apiResponse ApiResponse) error {
	return fmt.Errorf("telegram API error: %v, description: %v", apiResponse.ErrorCode, apiResponse.Description)
}
