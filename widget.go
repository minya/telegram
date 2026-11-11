package telegram

import "context"

// Widgets (Bot API 10.4): HTML blocks drawn inside a message, see
// https://core.telegram.org/bots/widgets. A widget's HTML is fixed once
// sent; its state is ChatData (per message) and PublicData (wherever it is
// shown), both changed only by the bot. A widget sent with AllowQueries asks
// the bot through WidgetQuery updates, answered with AnswerWidgetQuery.
//
// In the production environment widgets can only be sent to private chats.

// Widget describes the HTML widget of a message (Message.Widget).
type Widget struct {
	ID         string `json:"id,omitempty"` // set for a shared widget only
	Width      int    `json:"width,omitempty"`
	Height     int    `json:"height,omitempty"`
	HTML       string `json:"html"`
	PublicData string `json:"public_data,omitempty"`
	ChatData   string `json:"chat_data,omitempty"`
}

// Values of WidgetQueryType.Type.
const (
	WidgetQueryData            = "data"
	WidgetQuerySignData        = "sign_data"
	WidgetQuerySentTransaction = "sent_transaction"
)

// WidgetQueryType is the type of a widget query. Fields other than Type are
// set for the wallet types only and kept raw: this client does not model
// TON wallets.
type WidgetQueryType struct {
	Type       string         `json:"type"`
	WalletCard map[string]any `json:"wallet_card,omitempty"`
	Wallet     map[string]any `json:"wallet,omitempty"`
	SignData   string         `json:"sign_data,omitempty"`
	MsgHash    string         `json:"msg_hash,omitempty"`
}

// WidgetQuery is a request from a widget in a message sent by the bot or
// via the bot (Update.WidgetQuery). Answer it with AnswerWidgetQuery within
// 15 seconds. Data is the widget's untrusted input; From is authenticated by
// Telegram.
type WidgetQuery struct {
	ID              string           `json:"id"`
	From            *User            `json:"from"`
	Chat            *Chat            `json:"chat,omitempty"`
	MessageID       int64            `json:"message_id,omitempty"`
	InlineMessageID string           `json:"inline_message_id,omitempty"`
	ChatInstance    string           `json:"chat_instance,omitempty"`
	ChatType        string           `json:"chat_type,omitempty"`
	Type            *WidgetQueryType `json:"type,omitempty"`
	Data            string           `json:"data"`
	WidgetID        string           `json:"widget_id,omitempty"`
}

// SharedWidgetQuery is a request from a forwarded copy of a shared widget
// (Update.SharedWidgetQuery). It carries no chat or message.
type SharedWidgetQuery struct {
	ID       string           `json:"id"`
	From     *User            `json:"from"`
	Type     *WidgetQueryType `json:"type,omitempty"`
	Data     string           `json:"data"`
	WidgetID string           `json:"widget_id"`
}

// SendWidgetParams are the parameters of SendWidget. HTML is 1-32768 bytes
// and at most 4096 bytes after zlib level 9; PublicData and ChatData are up
// to 1024 bytes each. Width and Height are size hints in logical pixels, 0
// for automatic.
type SendWidgetParams struct {
	ChatID                int64            `json:"chat_id"`
	MessageThreadID       int64            `json:"message_thread_id,omitempty"`
	HTML                  string           `json:"html"`
	Width                 int              `json:"width,omitempty"`
	Height                int              `json:"height,omitempty"`
	PublicData            string           `json:"public_data,omitempty"`
	ChatData              string           `json:"chat_data,omitempty"`
	Caption               string           `json:"caption,omitempty"`
	ParseMode             string           `json:"parse_mode,omitempty"`
	CaptionEntities       []MessageEntity  `json:"caption_entities,omitempty"`
	ShowCaptionAboveMedia bool             `json:"show_caption_above_media,omitempty"`
	AllowQueries          bool             `json:"allow_queries,omitempty"`
	IsShared              bool             `json:"is_shared,omitempty"`
	DisableNotification   bool             `json:"disable_notification,omitempty"`
	ProtectContent        bool             `json:"protect_content,omitempty"`
	ReplyParameters       *ReplyParameters `json:"reply_parameters,omitempty"`
	ReplyMarkup           any              `json:"reply_markup,omitempty"`
}

// AnswerWidgetQueryParams answers a WidgetQuery or SharedWidgetQuery. Data
// (up to 1024 bytes) goes only to the widget that asked; it does not change
// the widget's state.
type AnswerWidgetQueryParams struct {
	WidgetQueryID string `json:"widget_query_id"`
	Data          string `json:"data,omitempty"`
}

// EditMessageWidgetDataParams changes the state of a widget in a message.
// A nil field is left as it is; a pointer to "" clears it.
type EditMessageWidgetDataParams struct {
	ChatID          int64   `json:"chat_id,omitempty"`
	MessageID       int64   `json:"message_id,omitempty"`
	InlineMessageID string  `json:"inline_message_id,omitempty"`
	PublicData      *string `json:"public_data,omitempty"`
	ChatData        *string `json:"chat_data,omitempty"`
}

// EditSharedWidgetDataParams changes the public data of a shared widget.
// CacheTime (5-3600 s) sets how long apps may cache it; 0 keeps the current
// value.
type EditSharedWidgetDataParams struct {
	WidgetID   string `json:"widget_id"`
	PublicData string `json:"public_data"`
	CacheTime  int    `json:"cache_time,omitempty"`
}

// SendSharedWidgetParams sends an existing shared widget again.
type SendSharedWidgetParams struct {
	ChatID                int64            `json:"chat_id"`
	MessageThreadID       int64            `json:"message_thread_id,omitempty"`
	WidgetID              string           `json:"widget_id"`
	ChatData              string           `json:"chat_data,omitempty"`
	Caption               string           `json:"caption,omitempty"`
	ParseMode             string           `json:"parse_mode,omitempty"`
	CaptionEntities       []MessageEntity  `json:"caption_entities,omitempty"`
	ShowCaptionAboveMedia bool             `json:"show_caption_above_media,omitempty"`
	DisableNotification   bool             `json:"disable_notification,omitempty"`
	ProtectContent        bool             `json:"protect_content,omitempty"`
	ReplyParameters       *ReplyParameters `json:"reply_parameters,omitempty"`
	ReplyMarkup           any              `json:"reply_markup,omitempty"`
}

// SendWidget sends a message with an HTML widget.
func (api *API) SendWidget(ctx context.Context, params SendWidgetParams) (*Message, error) {
	return api.callMessage(ctx, "sendWidget", params)
}

// SendSharedWidget sends a shared widget created by the bot again.
func (api *API) SendSharedWidget(ctx context.Context, params SendSharedWidgetParams) (*Message, error) {
	return api.callMessage(ctx, "sendSharedWidget", params)
}

// AnswerWidgetQuery answers a query from a widget.
func (api *API) AnswerWidgetQuery(ctx context.Context, params AnswerWidgetQueryParams) error {
	return api.call(ctx, "answerWidgetQuery", params, nil)
}

// EditMessageWidgetData changes the state of a widget in a message. Like
// editMessageText, an edit that changes nothing fails with "message is not
// modified".
func (api *API) EditMessageWidgetData(ctx context.Context, params EditMessageWidgetDataParams) error {
	return api.call(ctx, "editMessageWidgetData", params, nil)
}

// EditSharedWidgetData changes the public data of a shared widget, at most
// once per second.
func (api *API) EditSharedWidgetData(ctx context.Context, params EditSharedWidgetDataParams) error {
	return api.call(ctx, "editSharedWidgetData", params, nil)
}
