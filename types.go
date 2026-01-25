package telegram

type Update struct {
	UpdateId          int64          `json:"update_id"`
	Message           *Message       `json:"message,omitempty"`
	EditedMessage     *Message       `json:"edited_message,omitempty"`
	ChannelPost       *Message       `json:"channel_post,omitempty"`
	EditedChannelPost *Message       `json:"edited_channel_post,omitempty"`
	CallbackQuery     *CallbackQuery `json:"callback_query,omitempty"`
}

type Message struct {
	MessageId       int64           `json:"message_id"`
	From            *User           `json:"from,omitempty"`
	Date            int64           `json:"date"`
	Chat            Chat            `json:"chat"`
	Text            string          `json:"text,omitempty"`
	Entities        []MessageEntity `json:"entities,omitempty"`
	Caption         string          `json:"caption,omitempty"`
	CaptionEntities []MessageEntity `json:"caption_entities,omitempty"`
	Document        *Document       `json:"document,omitempty"`
	Photo           []PhotoSize     `json:"photo,omitempty"`
}

type Document struct {
	FileName     string    `json:"file_name"`
	MimeType     string    `json:"mime_type"`
	Thumb        PhotoSize `json:"thumb,omitempty"`
	FileID       string    `json:"file_id"`
	FileUniqueID string    `json:"file_unique_id"`
	FileSize     *int64    `json:"file_size,omitempty"`
}

type PhotoSize struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	FileSize     *int64 `json:"file_size,omitempty"`
	Width        uint16 `json:"width"`
	Height       uint16 `json:"height"`
}

type MessageEntity struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
	URL    string `json:"url,omitempty"`
	User   *User  `json:"user,omitempty"`
}

type User struct {
	Id           int64  `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name,omitempty"`
	UserName     string `json:"username,omitempty"`
	IsBot        bool   `json:"is_bot"`
	LanguageCode string `json:"language_code,omitempty"`
}

type Chat struct {
	Id                          int64  `json:"id"`
	Type                        string `json:"type"`
	Title                       string `json:"title,omitempty"`
	Username                    string `json:"username,omitempty"`
	FirstName                   string `json:"first_name,omitempty"`
	LastName                    string `json:"last_name,omitempty"`
	AllMembersAreAdministrators bool   `json:"all_members_are_administrators,omitempty"`
}

type ReplyMessage struct {
	ChatId      int64       `json:"chat_id"`
	Text        string      `json:"text"`
	ParseMode   string      `json:"parse_mode,omitempty"`
	ReplyMarkup interface{} `json:"reply_markup,omitempty"`
}

type ReplyDocument struct {
	ChatId      int64  `json:"chat_id"`
	Caption     string `json:"caption,omitempty"`
	ParseMode   string `json:"parse_mode,omitempty"`
	InputFile   InputFile
	ReplyMarkup interface{} `json:"reply_markup,omitempty"`
}

type InputFile struct {
	Content  []byte
	FileName string
}

type CallbackQuery struct {
	Id           string   `json:"id"`
	From         *User    `json:"from"`
	Data         string   `json:"data"`
	Message      *Message `json:"message,omitempty"`
	ChatInstance string   `json:"chat_instance,omitempty"`
}

type ReplyKeyboardMarkup struct {
	Keyboard        [][]KeyboardButton `json:"keyboard"`
	ResizeKeyboard  bool               `json:"resize_keyboard"`
	OneTimeKeyboard bool               `json:"one_time_keyboard"`
	Selective       bool               `json:"selective"`
}

type SetWebhookParams struct {
	Url                string     `json:"url"`
	Certificate        *InputFile `json:"certificate,omitempty"`
	IpAddress          string     `json:"ip_address,omitempty"`
	MaxConnections     int        `json:"max_connections,omitempty"`
	AllowedUpdates     []string   `json:"allowed_updates,omitempty"`
	DropPendingUpdates bool       `json:"drop_pending_updates,omitempty"`
	SecretToken        string     `json:"secret_token,omitempty"`
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type KeyboardButton struct {
	Text            string `json:"text"`
	RequestContact  bool   `json:"request_contact"`
	RequestLocation bool   `json:"request_location"`
}

type InlineKeyboardButton struct {
	Text         string `json:"text"`
	Url          string `json:"url,omitempty"`
	CallbackData string `json:"callback_data,omitempty"`
	WebApp      *WebAppInfo `json:"web_app,omitempty"`
}

type WebAppInfo struct {
	Url string `json:"url"`
}

type File struct {
	FileID       string  `json:"file_id"`
	FileUniqueID string  `json:"file_unique_id"`
	FileSize     *int64  `json:"file_size,omitempty"`
	FilePath     *string `json:"file_path"`
}

type AnswerCallbackQueryParams struct {
	CallbackQueryID string  `json:"callback_query_id"`
	Text            string  `json:"text,omitempty"`
	ShowAlert       bool    `json:"show_alert,omitempty"`
	Url             *string `json:"url,omitempty"`
	CacheTime       int     `json:"cache_time,omitempty"`
}

type EditMessageTextParams struct {
	ChatID      int64       `json:"chat_id,omitempty"`
	MessageID   int64       `json:"message_id,omitempty"`
	Text        string      `json:"text"`
	ParseMode   string      `json:"parse_mode,omitempty"`
	ReplyMarkup interface{} `json:"reply_markup,omitempty"`
}

type SetChatMenuButtonParams struct {
	ChatID     int64       `json:"chat_id,omitempty"`
	MenuButton *MenuButton `json:"menu_button,omitempty"`
}

type MenuButton struct {
	Type   string      `json:"type"`
	Text   string      `json:"text,omitempty"`
	WebApp *WebAppInfo `json:"web_app,omitempty"`
}

func (msg *Message) HasDocument() bool {
	if msg == nil || msg.Document == nil {
		return false
	}
	return msg.Document.FileID != ""
}
