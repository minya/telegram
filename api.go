package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/minya/goutils/web"
)

// DefaultBaseURL is the public Bot API endpoint.
const DefaultBaseURL = "https://api.telegram.org"

// API is a Telegram Bot API client. It is safe for concurrent use.
type API struct {
	token   string
	baseURL string
	client  *http.Client
}

// Option configures an API created by NewAPI.
type Option func(*API)

// WithBaseURL points the client at a different Bot API server, such as a
// local telegram-bot-api instance or a test server.
func WithBaseURL(baseURL string) Option {
	return func(api *API) { api.baseURL = strings.TrimRight(baseURL, "/") }
}

// WithHTTPClient replaces the default HTTP client. Its timeout must be longer
// than the long-poll timeout used by Poll.
func WithHTTPClient(client *http.Client) Option {
	return func(api *API) { api.client = client }
}

// NewAPI creates a client for the bot with the given token.
func NewAPI(token string, opts ...Option) *API {
	api := &API{
		token:   token,
		baseURL: DefaultBaseURL,
		client: &http.Client{
			Transport: web.DefaultTransport(1000),
			Timeout:   60 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(api)
	}
	return api
}

type apiResponse struct {
	Ok          bool                `json:"ok"`
	Description string              `json:"description,omitempty"`
	ErrorCode   int                 `json:"error_code,omitempty"`
	Result      json.RawMessage     `json:"result,omitempty"`
	Parameters  *ResponseParameters `json:"parameters,omitempty"`
}

// GetMe returns basic information about the bot.
func (api *API) GetMe(ctx context.Context) (*User, error) {
	var user User
	if err := api.call(ctx, "getMe", nil, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUpdates receives incoming updates using long polling.
func (api *API) GetUpdates(ctx context.Context, params GetUpdatesParams) ([]Update, error) {
	var updates []Update
	if err := api.call(ctx, "getUpdates", params, &updates); err != nil {
		return nil, err
	}
	return updates, nil
}

// SendMessage sends a text message and returns it as delivered.
func (api *API) SendMessage(ctx context.Context, params SendMessageParams) (*Message, error) {
	return api.callMessage(ctx, "sendMessage", params)
}

// SendDocument sends a general file, either uploaded from InputFile.Content
// or referenced by InputFile.Ref.
func (api *API) SendDocument(ctx context.Context, params SendDocumentParams) (*Message, error) {
	var msg Message
	if err := api.callWithFile(ctx, "sendDocument", "document", params.Document, params, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// SendPhoto sends a photo, either uploaded from InputFile.Content or
// referenced by InputFile.Ref.
func (api *API) SendPhoto(ctx context.Context, params SendPhotoParams) (*Message, error) {
	var msg Message
	if err := api.callWithFile(ctx, "sendPhoto", "photo", params.Photo, params, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// ForwardMessage forwards a message.
func (api *API) ForwardMessage(ctx context.Context, params ForwardMessageParams) (*Message, error) {
	return api.callMessage(ctx, "forwardMessage", params)
}

// CopyMessage copies a message without a link to the original and returns
// the ID of the new message.
func (api *API) CopyMessage(ctx context.Context, params CopyMessageParams) (int64, error) {
	var id struct {
		MessageID int64 `json:"message_id"`
	}
	if err := api.call(ctx, "copyMessage", params, &id); err != nil {
		return 0, err
	}
	return id.MessageID, nil
}

// EditMessageText edits the text of a message. For inline messages Telegram
// returns no message, so the result is nil.
func (api *API) EditMessageText(ctx context.Context, params EditMessageTextParams) (*Message, error) {
	return api.callEdit(ctx, "editMessageText", params)
}

// EditMessageReplyMarkup edits only the reply markup of a message. For inline
// messages the result is nil.
func (api *API) EditMessageReplyMarkup(ctx context.Context, params EditMessageReplyMarkupParams) (*Message, error) {
	return api.callEdit(ctx, "editMessageReplyMarkup", params)
}

// DeleteMessage deletes a message.
func (api *API) DeleteMessage(ctx context.Context, chatID, messageID int64) error {
	params := struct {
		ChatID    int64 `json:"chat_id"`
		MessageID int64 `json:"message_id"`
	}{chatID, messageID}
	return api.call(ctx, "deleteMessage", params, nil)
}

// SendChatAction shows a status such as ChatActionTyping in the chat.
func (api *API) SendChatAction(ctx context.Context, chatID int64, action string) error {
	params := struct {
		ChatID int64  `json:"chat_id"`
		Action string `json:"action"`
	}{chatID, action}
	return api.call(ctx, "sendChatAction", params, nil)
}

// AnswerCallbackQuery answers a callback query from an inline keyboard.
func (api *API) AnswerCallbackQuery(ctx context.Context, params AnswerCallbackQueryParams) error {
	return api.call(ctx, "answerCallbackQuery", params, nil)
}

// SetMyCommands changes the list of the bot's commands.
func (api *API) SetMyCommands(ctx context.Context, params SetMyCommandsParams) error {
	return api.call(ctx, "setMyCommands", params, nil)
}

// SetChatMenuButton changes the bot's menu button in a chat, or the default
// menu button when ChatID is zero.
func (api *API) SetChatMenuButton(ctx context.Context, params SetChatMenuButtonParams) error {
	return api.call(ctx, "setChatMenuButton", params, nil)
}

// SetWebhook registers a webhook URL. A self-signed Certificate, if set, is
// uploaded with the request.
func (api *API) SetWebhook(ctx context.Context, params SetWebhookParams) error {
	if params.Certificate == nil {
		return api.call(ctx, "setWebhook", params, nil)
	}
	return api.callWithFile(ctx, "setWebhook", "certificate", *params.Certificate, params, nil)
}

// DeleteWebhook removes the webhook so that GetUpdates can be used again.
func (api *API) DeleteWebhook(ctx context.Context, dropPendingUpdates bool) error {
	params := struct {
		DropPendingUpdates bool `json:"drop_pending_updates,omitempty"`
	}{dropPendingUpdates}
	return api.call(ctx, "deleteWebhook", params, nil)
}

// GetWebhookInfo returns the current webhook status.
func (api *API) GetWebhookInfo(ctx context.Context) (*WebhookInfo, error) {
	var info WebhookInfo
	if err := api.call(ctx, "getWebhookInfo", nil, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// GetFile returns a File whose FilePath can be passed to DownloadFile.
func (api *API) GetFile(ctx context.Context, fileID string) (*File, error) {
	params := struct {
		FileID string `json:"file_id"`
	}{fileID}
	var file File
	if err := api.call(ctx, "getFile", params, &file); err != nil {
		return nil, err
	}
	return &file, nil
}

// DownloadFile downloads the contents of a file returned by GetFile.
func (api *API) DownloadFile(ctx context.Context, file *File) ([]byte, error) {
	if file == nil || file.FilePath == "" {
		return nil, fmt.Errorf("telegram: download file: no file path (files over 20 MB cannot be downloaded)")
	}
	url := fmt.Sprintf("%s/file/bot%s/%s", api.baseURL, api.token, file.FilePath)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, api.redact(err)
	}
	resp, err := api.client.Do(req)
	if err != nil {
		return nil, api.redact(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("telegram: download file: HTTP %d", resp.StatusCode)
	}
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("telegram: download file: %w", api.redact(err))
	}
	return content, nil
}

func (api *API) methodURL(method string) string {
	return fmt.Sprintf("%s/bot%s/%s", api.baseURL, api.token, method)
}

// call invokes a method with a JSON payload and decodes the result into
// result, which may be nil when the result is not needed.
func (api *API) call(ctx context.Context, method string, params any, result any) error {
	var body io.Reader = http.NoBody
	if params != nil {
		payload, err := json.Marshal(params)
		if err != nil {
			return fmt.Errorf("telegram: %s: encode params: %w", method, err)
		}
		body = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, api.methodURL(method), body)
	if err != nil {
		return api.redact(err)
	}
	req.Header.Set("Content-Type", "application/json")
	return api.do(req, method, result)
}

// callWithFile invokes a method as multipart/form-data. If file has Content it
// is uploaded under field; otherwise file.Ref (a file_id or URL) is sent.
func (api *API) callWithFile(ctx context.Context, method, field string, file InputFile, params any, result any) error {
	fields, err := formFields(params)
	if err != nil {
		return fmt.Errorf("telegram: %s: encode params: %w", method, err)
	}
	if file.Content == nil {
		if file.Ref == "" {
			return fmt.Errorf("telegram: %s: %s has neither Content nor Ref", method, field)
		}
		fields[field] = file.Ref
	}

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for name, value := range fields {
		if err := mw.WriteField(name, value); err != nil {
			return fmt.Errorf("telegram: %s: write form: %w", method, err)
		}
	}
	if file.Content != nil {
		fw, err := mw.CreateFormFile(field, file.FileName)
		if err != nil {
			return fmt.Errorf("telegram: %s: write form: %w", method, err)
		}
		if _, err := fw.Write(file.Content); err != nil {
			return fmt.Errorf("telegram: %s: write form: %w", method, err)
		}
	}
	if err := mw.Close(); err != nil {
		return fmt.Errorf("telegram: %s: write form: %w", method, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, api.methodURL(method), &buf)
	if err != nil {
		return api.redact(err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return api.do(req, method, result)
}

func (api *API) do(req *http.Request, method string, result any) error {
	resp, err := api.client.Do(req)
	if err != nil {
		return api.redact(err)
	}
	defer resp.Body.Close()

	var apiResp apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("telegram: %s: decode response (HTTP %d): %w", method, resp.StatusCode, api.redact(err))
	}
	if !apiResp.Ok {
		apiErr := &Error{
			Method:      method,
			Code:        apiResp.ErrorCode,
			Description: apiResp.Description,
		}
		if apiResp.Parameters != nil {
			apiErr.RetryAfter = apiResp.Parameters.RetryAfter
			apiErr.MigrateToChatID = apiResp.Parameters.MigrateToChatID
		}
		return apiErr
	}
	if result == nil || len(apiResp.Result) == 0 {
		return nil
	}
	if err := json.Unmarshal(apiResp.Result, result); err != nil {
		return fmt.Errorf("telegram: %s: decode result: %w", method, err)
	}
	return nil
}

func (api *API) callMessage(ctx context.Context, method string, params any) (*Message, error) {
	var msg Message
	if err := api.call(ctx, method, params, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// callEdit handles edit methods, which return the edited Message for regular
// messages and plain true for inline messages.
func (api *API) callEdit(ctx context.Context, method string, params any) (*Message, error) {
	var raw json.RawMessage
	if err := api.call(ctx, method, params, &raw); err != nil {
		return nil, err
	}
	if string(raw) == "true" {
		return nil, nil
	}
	var msg Message
	if err := json.Unmarshal(raw, &msg); err != nil {
		return nil, fmt.Errorf("telegram: %s: decode result: %w", method, err)
	}
	return &msg, nil
}

// formFields flattens params into multipart form values using their JSON
// field names: strings are sent as-is, everything else as JSON.
func formFields(params any) (map[string]string, error) {
	payload, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, err
	}
	fields := make(map[string]string, len(raw))
	for name, value := range raw {
		var s string
		if json.Unmarshal(value, &s) == nil {
			fields[name] = s
		} else {
			fields[name] = string(value)
		}
	}
	return fields, nil
}
