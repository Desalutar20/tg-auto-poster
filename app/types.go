package app

type update struct {
	ID            int            `json:"update_id"`
	Message       *message       `json:"message,omitempty"`
	ChannelPost   *message       `json:"channel_post,omitempty"`
	CallbackQuery *callbackQuery `json:"callback_query,omitempty"`
}

type callbackQuery struct {
	ID      string   `json:"id"`
	From    user     `json:"from"`
	Message *message `json:"message"`
	Data    string   `json:"data"`
}

type baseResponse[T any] struct {
	Ok          bool    `json:"ok"`
	Description *string `json:"description"`
	Result      T       `json:"result"`
}

type user struct {
	ID int64 `json:"id"`
}

type messageOriginChannel struct {
	Type            string `json:"type"`
	Date            int64  `json:"date"`
	Chat            chat   `json:"chat"`
	MessageID       int    `json:"message_id"`
	AuthorSignature string `json:"author_signature,omitempty"`
}

type messageEntity struct {
	Type          string  `json:"type"`
	Offset        int     `json:"offset"`
	Length        int     `json:"length"`
	Url           *string `json:"url,omitempty"`
	User          *user   `json:"user,omitempty"`
	Language      *string `json:"language,omitempty"`
	CustomEmojiID *string `json:"custom_emoji_id,omitempty"`
}

type photo struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	Width        int64  `json:"width"`
	Height       int64  `json:"height"`
	FileSize     *int64 `json:"file_size,omitempty"`
}

type message struct {
	ID              int64                 `json:"message_id"`
	Chat            chat                  `json:"chat"`
	Text            string                `json:"text"`
	From            *user                 `json:"from,omitempty"`
	ForwardOrigin   *messageOriginChannel `json:"forward_origin,omitempty"`
	Entities        []messageEntity       `json:"entities"`
	Photo           []photo               `json:"photo"`
	Caption         *string               `json:"caption,omitempty"`
	CaptionEntities []messageEntity       `json:"caption_entities"`
}

type chat struct {
	ID int64 `json:"id"`
}

type callbackAnwser struct {
	ID        string `json:"callback_query_id"`
	Text      string `json:"text"`
	ShowAlert bool   `json:"show_alert"`
}

type inlineKeyboardMarkup struct {
	Text         string   `json:"text"`
	Url          string   `json:"url"`
	CallbackData Callback `json:"callback_data"`
}

type sendMessageRequest struct {
	ChatID      int64  `json:"chat_id"`
	Text        string `json:"text"`
	ParseMode   string `json:"parse_mode"`
	ReplyMarkup struct {
		InlineKeyboard [][]inlineKeyboardMarkup `json:"inline_keyboard,omitempty"`
	} `json:"reply_markup" `
}

type sendPhotoRequest struct {
	ChatID    int64  `json:"chat_id"`
	Photo     string `json:"photo"`
	Caption   string `json:"caption,omitempty"`
	ParseMode string `json:"parse_mode,omitempty"`
}

type copyMessageRequest struct {
	ChatID     int64  `json:"chat_id"`
	FromChatID int64  `json:"from_chat_id"`
	MessageID  int    `json:"message_id"`
	ParseMode  string `json:"parse_mode"`
}

type deleteMessageRequest struct {
	ChatID    int64 `json:"chat_id"`
	MessageID int64 `json:"message_id"`
}

type pinMessageRequest struct {
	ChatID    int64 `json:"chat_id"`
	MessageID int64 `json:"message_id"`
}
