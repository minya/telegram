package telegram

// Values for the ParseMode fields.
const (
	ParseModeHTML       = "HTML"
	ParseModeMarkdownV2 = "MarkdownV2"
	ParseModeMarkdown   = "Markdown" // legacy mode, prefer MarkdownV2
)

// Values of Chat.Type.
const (
	ChatTypePrivate    = "private"
	ChatTypeGroup      = "group"
	ChatTypeSupergroup = "supergroup"
	ChatTypeChannel    = "channel"
)

// Actions for SendChatAction.
const (
	ChatActionTyping         = "typing"
	ChatActionUploadPhoto    = "upload_photo"
	ChatActionUploadVideo    = "upload_video"
	ChatActionUploadVoice    = "upload_voice"
	ChatActionUploadDocument = "upload_document"
	ChatActionRecordVoice    = "record_voice"
	ChatActionFindLocation   = "find_location"
)

type Update struct {
	UpdateID          int64              `json:"update_id"`
	Message           *Message           `json:"message,omitempty"`
	EditedMessage     *Message           `json:"edited_message,omitempty"`
	ChannelPost       *Message           `json:"channel_post,omitempty"`
	EditedChannelPost *Message           `json:"edited_channel_post,omitempty"`
	InlineQuery       *InlineQuery       `json:"inline_query,omitempty"`
	CallbackQuery     *CallbackQuery     `json:"callback_query,omitempty"`
	MyChatMember      *ChatMemberUpdated `json:"my_chat_member,omitempty"`
	ChatMember        *ChatMemberUpdated `json:"chat_member,omitempty"`
}

type Message struct {
	MessageID       int64           `json:"message_id"`
	MessageThreadID int64           `json:"message_thread_id,omitempty"`
	From            *User           `json:"from,omitempty"`
	SenderChat      *Chat           `json:"sender_chat,omitempty"`
	Date            int64           `json:"date"`
	Chat            Chat            `json:"chat"`
	ReplyToMessage  *Message        `json:"reply_to_message,omitempty"`
	Text            string          `json:"text,omitempty"`
	Entities        []MessageEntity `json:"entities,omitempty"`
	Caption         string          `json:"caption,omitempty"`
	CaptionEntities []MessageEntity `json:"caption_entities,omitempty"`
	Document        *Document       `json:"document,omitempty"`
	Photo           []PhotoSize     `json:"photo,omitempty"`
	Audio           *Audio          `json:"audio,omitempty"`
	Video           *Video          `json:"video,omitempty"`
	Voice           *Voice          `json:"voice,omitempty"`
	Sticker         *Sticker        `json:"sticker,omitempty"`
	Contact         *Contact        `json:"contact,omitempty"`
	Location        *Location       `json:"location,omitempty"`
}

func (msg *Message) HasDocument() bool {
	if msg == nil || msg.Document == nil {
		return false
	}
	return msg.Document.FileID != ""
}

type MessageEntity struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
	URL    string `json:"url,omitempty"`
	User   *User  `json:"user,omitempty"`
}

type User struct {
	ID           int64  `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name,omitempty"`
	Username     string `json:"username,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
}

type Chat struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title,omitempty"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	IsForum   bool   `json:"is_forum,omitempty"`
}

type ChatMember struct {
	Status string `json:"status"`
	User   User   `json:"user"`
}

type ChatMemberUpdated struct {
	Chat          Chat       `json:"chat"`
	From          User       `json:"from"`
	Date          int64      `json:"date"`
	OldChatMember ChatMember `json:"old_chat_member"`
	NewChatMember ChatMember `json:"new_chat_member"`
}

type PhotoSize struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	FileSize     *int64 `json:"file_size,omitempty"`
}

type Document struct {
	FileID       string     `json:"file_id"`
	FileUniqueID string     `json:"file_unique_id"`
	Thumbnail    *PhotoSize `json:"thumbnail,omitempty"`
	FileName     string     `json:"file_name,omitempty"`
	MimeType     string     `json:"mime_type,omitempty"`
	FileSize     *int64     `json:"file_size,omitempty"`
}

type Audio struct {
	FileID       string     `json:"file_id"`
	FileUniqueID string     `json:"file_unique_id"`
	Duration     int        `json:"duration"`
	Performer    string     `json:"performer,omitempty"`
	Title        string     `json:"title,omitempty"`
	FileName     string     `json:"file_name,omitempty"`
	MimeType     string     `json:"mime_type,omitempty"`
	FileSize     *int64     `json:"file_size,omitempty"`
	Thumbnail    *PhotoSize `json:"thumbnail,omitempty"`
}

type Video struct {
	FileID       string     `json:"file_id"`
	FileUniqueID string     `json:"file_unique_id"`
	Width        int        `json:"width"`
	Height       int        `json:"height"`
	Duration     int        `json:"duration"`
	Thumbnail    *PhotoSize `json:"thumbnail,omitempty"`
	FileName     string     `json:"file_name,omitempty"`
	MimeType     string     `json:"mime_type,omitempty"`
	FileSize     *int64     `json:"file_size,omitempty"`
}

type Voice struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	Duration     int    `json:"duration"`
	MimeType     string `json:"mime_type,omitempty"`
	FileSize     *int64 `json:"file_size,omitempty"`
}

type Sticker struct {
	FileID       string     `json:"file_id"`
	FileUniqueID string     `json:"file_unique_id"`
	Type         string     `json:"type"`
	Width        int        `json:"width"`
	Height       int        `json:"height"`
	IsAnimated   bool       `json:"is_animated"`
	IsVideo      bool       `json:"is_video"`
	Thumbnail    *PhotoSize `json:"thumbnail,omitempty"`
	Emoji        string     `json:"emoji,omitempty"`
	SetName      string     `json:"set_name,omitempty"`
	FileSize     *int64     `json:"file_size,omitempty"`
}

type Contact struct {
	PhoneNumber string `json:"phone_number"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name,omitempty"`
	UserID      int64  `json:"user_id,omitempty"`
	VCard       string `json:"vcard,omitempty"`
}

type Location struct {
	Latitude           float64 `json:"latitude"`
	Longitude          float64 `json:"longitude"`
	HorizontalAccuracy float64 `json:"horizontal_accuracy,omitempty"`
}

type File struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	FileSize     *int64 `json:"file_size,omitempty"`
	FilePath     string `json:"file_path,omitempty"` // empty for files over 20 MB
}

// InputFile is a file to send. Set Content and FileName to upload new data,
// or Ref to a file_id or HTTP URL to send an existing file.
type InputFile struct {
	FileName string
	Content  []byte
	Ref      string
}

type InlineQuery struct {
	ID       string `json:"id"`
	From     *User  `json:"from"`
	Query    string `json:"query"`
	Offset   string `json:"offset"`
	ChatType string `json:"chat_type,omitempty"`
}

type CallbackQuery struct {
	ID              string   `json:"id"`
	From            *User    `json:"from"`
	Message         *Message `json:"message,omitempty"`
	InlineMessageID string   `json:"inline_message_id,omitempty"`
	ChatInstance    string   `json:"chat_instance,omitempty"`
	Data            string   `json:"data,omitempty"`
}

// Reply markup. Any of these can be used as a ReplyMarkup value.

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type InlineKeyboardButton struct {
	Text         string      `json:"text"`
	URL          string      `json:"url,omitempty"`
	CallbackData string      `json:"callback_data,omitempty"`
	WebApp       *WebAppInfo `json:"web_app,omitempty"`
}

type ReplyKeyboardMarkup struct {
	Keyboard              [][]KeyboardButton `json:"keyboard"`
	IsPersistent          bool               `json:"is_persistent,omitempty"`
	ResizeKeyboard        bool               `json:"resize_keyboard,omitempty"`
	OneTimeKeyboard       bool               `json:"one_time_keyboard,omitempty"`
	InputFieldPlaceholder string             `json:"input_field_placeholder,omitempty"`
	Selective             bool               `json:"selective,omitempty"`
}

type KeyboardButton struct {
	Text            string      `json:"text"`
	RequestContact  bool        `json:"request_contact,omitempty"`
	RequestLocation bool        `json:"request_location,omitempty"`
	WebApp          *WebAppInfo `json:"web_app,omitempty"`
}

// ReplyKeyboardRemove hides a custom keyboard. RemoveKeyboard must be true.
type ReplyKeyboardRemove struct {
	RemoveKeyboard bool `json:"remove_keyboard"`
	Selective      bool `json:"selective,omitempty"`
}

// ForceReply makes the client show a reply interface. ForceReply must be true.
type ForceReply struct {
	ForceReply            bool   `json:"force_reply"`
	InputFieldPlaceholder string `json:"input_field_placeholder,omitempty"`
	Selective             bool   `json:"selective,omitempty"`
}

type WebAppInfo struct {
	URL string `json:"url"`
}

type MenuButton struct {
	Type   string      `json:"type"`
	Text   string      `json:"text,omitempty"`
	WebApp *WebAppInfo `json:"web_app,omitempty"`
}

type BotCommand struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

type ReplyParameters struct {
	MessageID                int64 `json:"message_id"`
	ChatID                   int64 `json:"chat_id,omitempty"`
	AllowSendingWithoutReply bool  `json:"allow_sending_without_reply,omitempty"`
}

type LinkPreviewOptions struct {
	IsDisabled       bool   `json:"is_disabled,omitempty"`
	URL              string `json:"url,omitempty"`
	PreferSmallMedia bool   `json:"prefer_small_media,omitempty"`
	PreferLargeMedia bool   `json:"prefer_large_media,omitempty"`
	ShowAboveText    bool   `json:"show_above_text,omitempty"`
}

type WebhookInfo struct {
	URL                          string   `json:"url"`
	HasCustomCertificate         bool     `json:"has_custom_certificate"`
	PendingUpdateCount           int      `json:"pending_update_count"`
	IPAddress                    string   `json:"ip_address,omitempty"`
	LastErrorDate                int64    `json:"last_error_date,omitempty"`
	LastErrorMessage             string   `json:"last_error_message,omitempty"`
	LastSynchronizationErrorDate int64    `json:"last_synchronization_error_date,omitempty"`
	MaxConnections               int      `json:"max_connections,omitempty"`
	AllowedUpdates               []string `json:"allowed_updates,omitempty"`
}

// Method parameters.

type GetUpdatesParams struct {
	Offset         int64    `json:"offset,omitempty"`
	Limit          int      `json:"limit,omitempty"`
	Timeout        int      `json:"timeout,omitempty"` // long-poll timeout in seconds
	AllowedUpdates []string `json:"allowed_updates,omitempty"`
}

type SendMessageParams struct {
	ChatID              int64               `json:"chat_id"`
	MessageThreadID     int64               `json:"message_thread_id,omitempty"`
	Text                string              `json:"text"`
	ParseMode           string              `json:"parse_mode,omitempty"`
	LinkPreviewOptions  *LinkPreviewOptions `json:"link_preview_options,omitempty"`
	DisableNotification bool                `json:"disable_notification,omitempty"`
	ProtectContent      bool                `json:"protect_content,omitempty"`
	ReplyParameters     *ReplyParameters    `json:"reply_parameters,omitempty"`
	ReplyMarkup         any                 `json:"reply_markup,omitempty"`
}

type SendDocumentParams struct {
	ChatID              int64            `json:"chat_id"`
	MessageThreadID     int64            `json:"message_thread_id,omitempty"`
	Document            InputFile        `json:"-"`
	Caption             string           `json:"caption,omitempty"`
	ParseMode           string           `json:"parse_mode,omitempty"`
	DisableNotification bool             `json:"disable_notification,omitempty"`
	ProtectContent      bool             `json:"protect_content,omitempty"`
	ReplyParameters     *ReplyParameters `json:"reply_parameters,omitempty"`
	ReplyMarkup         any              `json:"reply_markup,omitempty"`
}

type SendPhotoParams struct {
	ChatID              int64            `json:"chat_id"`
	MessageThreadID     int64            `json:"message_thread_id,omitempty"`
	Photo               InputFile        `json:"-"`
	Caption             string           `json:"caption,omitempty"`
	ParseMode           string           `json:"parse_mode,omitempty"`
	DisableNotification bool             `json:"disable_notification,omitempty"`
	ProtectContent      bool             `json:"protect_content,omitempty"`
	ReplyParameters     *ReplyParameters `json:"reply_parameters,omitempty"`
	ReplyMarkup         any              `json:"reply_markup,omitempty"`
}

type ForwardMessageParams struct {
	ChatID              int64 `json:"chat_id"`
	MessageThreadID     int64 `json:"message_thread_id,omitempty"`
	FromChatID          int64 `json:"from_chat_id"`
	MessageID           int64 `json:"message_id"`
	DisableNotification bool  `json:"disable_notification,omitempty"`
	ProtectContent      bool  `json:"protect_content,omitempty"`
}

type CopyMessageParams struct {
	ChatID              int64            `json:"chat_id"`
	MessageThreadID     int64            `json:"message_thread_id,omitempty"`
	FromChatID          int64            `json:"from_chat_id"`
	MessageID           int64            `json:"message_id"`
	Caption             string           `json:"caption,omitempty"`
	ParseMode           string           `json:"parse_mode,omitempty"`
	DisableNotification bool             `json:"disable_notification,omitempty"`
	ProtectContent      bool             `json:"protect_content,omitempty"`
	ReplyParameters     *ReplyParameters `json:"reply_parameters,omitempty"`
	ReplyMarkup         any              `json:"reply_markup,omitempty"`
}

// EditMessageTextParams identifies the message either by ChatID and
// MessageID, or by InlineMessageID.
type EditMessageTextParams struct {
	ChatID             int64               `json:"chat_id,omitempty"`
	MessageID          int64               `json:"message_id,omitempty"`
	InlineMessageID    string              `json:"inline_message_id,omitempty"`
	Text               string              `json:"text"`
	ParseMode          string              `json:"parse_mode,omitempty"`
	LinkPreviewOptions *LinkPreviewOptions `json:"link_preview_options,omitempty"`
	ReplyMarkup        any                 `json:"reply_markup,omitempty"`
}

type EditMessageReplyMarkupParams struct {
	ChatID          int64  `json:"chat_id,omitempty"`
	MessageID       int64  `json:"message_id,omitempty"`
	InlineMessageID string `json:"inline_message_id,omitempty"`
	ReplyMarkup     any    `json:"reply_markup,omitempty"`
}

type AnswerCallbackQueryParams struct {
	CallbackQueryID string `json:"callback_query_id"`
	Text            string `json:"text,omitempty"`
	ShowAlert       bool   `json:"show_alert,omitempty"`
	URL             string `json:"url,omitempty"`
	CacheTime       int    `json:"cache_time,omitempty"`
}

type SetMyCommandsParams struct {
	Commands     []BotCommand `json:"commands"`
	Scope        any          `json:"scope,omitempty"`
	LanguageCode string       `json:"language_code,omitempty"`
}

type SetChatMenuButtonParams struct {
	ChatID     int64       `json:"chat_id,omitempty"`
	MenuButton *MenuButton `json:"menu_button,omitempty"`
}

type SetWebhookParams struct {
	URL                string     `json:"url"`
	Certificate        *InputFile `json:"-"` // public key of a self-signed certificate
	IPAddress          string     `json:"ip_address,omitempty"`
	MaxConnections     int        `json:"max_connections,omitempty"`
	AllowedUpdates     []string   `json:"allowed_updates,omitempty"`
	DropPendingUpdates bool       `json:"drop_pending_updates,omitempty"`
	SecretToken        string     `json:"secret_token,omitempty"`
}
