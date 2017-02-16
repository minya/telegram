package telegram

type Update struct {
	UpdateId int     `json:update_id`
	Message  Message `json:message`
}

type Message struct {
	MessageId int    `json:message_id`
	From      User   `json:from`
	Date      int    `json:date`
	Chat      Chat   `json:chat`
	Text      string `json:text`
}

type User struct {
	Id        int    `json:id`
	FirstName string `json:first_name`
	LastName  string `json:last_name`
	UserName  string `json:username`
}

type Chat struct {
	Id                          int    `json:id`
	Type                        string `json:chat_type`
	Title                       string `json:title`
	Username                    string `json:username`
	FirstName                   string `json:first_name`
	LastName                    string `json:last_name`
	AllMembersAreAdministrators bool   `json:all_members_are_administrators`
}
