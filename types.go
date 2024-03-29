package telegram

type Update struct {
	UpdateId      int           `json:"update_id"`
	Message       Message       `json:"message,omitempty"`
	CallbackQuery CallbackQuery `json:"callback_query,omitempty"`
}

type Message struct {
	MessageId int      `json:"message_id"`
	From      User     `json:"from"`
	Date      int64    `json:"date"`
	Chat      Chat     `json:"chat"`
	Text      string   `json:"text,omitempty"`
	Document  Document `json:"document,omitempty"`
}

type Document struct {
	FileName     string    `json:"file_name"`
	MimeType     string    `json:"mime_type"`
	Thumb        PhotoSize `json:"thumb"`
	FileID       string    `json:"file_id"`
	FileUniqueID string    `json:"file_unique_id"`
	FileSize     uint32    `json:"file_size"`
}

type PhotoSize struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	FileSize     uint32 `json:"file_size"`
	Width        uint16 `json:"width"`
	Height       uint16 `json:"height"`
}

type User struct {
	Id           int    `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	UserName     string `json:"username"`
	IsBot        bool   `json:"is_bot"`
	LanguageCode string `json:"language_code"`
}

type Chat struct {
	Id                          int    `json:"id"`
	Type                        string `json:"chat_type"`
	Title                       string `json:"title"`
	Username                    string `json:"username"`
	FirstName                   string `json:"first_name"`
	LastName                    string `json:"last_name"`
	AllMembersAreAdministrators bool   `json:"all_members_are_administrators"`
}

type ReplyMessage struct {
	ChatId      int         `json:"chat_id"`
	Text        string      `json:"text"`
	ParseMode   string      `json:"parse_mode,omitempty"`
	ReplyMarkup interface{} `json:"reply_markup,omitempty"`
}

type ReplyDocument struct {
	ChatId      int    `json:"chat_id"`
	Caption     string `json:"parse_mode,omitempty"`
	InputFile   InputFile
	ReplyMarkup interface{} `json:"reply_markup,omitempty"`
}

type InputFile struct {
	Content  []byte
	FileName string
}

type CallbackQuery struct {
	Id           string  `json:"id"`
	From         User    `json:"from"`
	Data         string  `json:"data"`
	Message      Message `json:"message,omitempty"`
	ChatInstance string  `json:"chat_instance,omitempty"`
}

type ReplyKeyboardMarkup struct {
	Keyboard        [][]KeyboardButton `json:"keyboard"`
	ResizeKeyboard  bool               `json:"resize_keyboard"`
	OneTimeKeyboard bool               `json:"one_time_keyboard"`
	Selective       bool               `json:"selective"`
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
}

type File struct {
	FileID       string  `json:"file_id"`
	FileUniqueID string  `json:"file_unique_id"`
	FileSize     *int    `json:"file_size"`
	FilePath     *string `json:"file_path"`
}

func (msg *Message) HasDocument() bool {
	return msg.Document.FileID != ""
}
