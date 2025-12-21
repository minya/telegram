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

func (api *Api) GetUpdates(offset int64) ([]Update, error) {
	type getUpdatesParams struct {
		Offset  int64  `json:"offset"`
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
	return io.ReadAll(resp.Body)
}

func (api *Api) AnswerCallbackQuery(params *AnswerCallbackQueryParams) error {
	_, err := api.callMethod("answerCallbackQuery", params)
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
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return responseBytes, err
}

type UpdatesResult struct {
	Ok     bool     `json:"ok"`
	Result []Update `json:"result,omitempty"`
}
