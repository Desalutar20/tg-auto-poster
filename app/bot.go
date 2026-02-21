package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
)

const TELEGRAM_BOT_BASE_URL string = "https://api.telegram.org/bot"

type Callback string

const NONE_DATA = "none"
const START_CALLBACK_DATA Callback = "start"
const STOP_CALLBACK_DATA Callback = "stop"
const ADD_CHAT_DATA Callback = "add-chat"
const RESET_CHATS_DATA Callback = "reset-chats"
const CHOOSE_INTERVAL_DATA Callback = "choose-interval"
const CHANGE_MESSAGE Callback = "change-message"
const PIN_DATA Callback = "pin"
const REMOVE_LAST_DATA Callback = "remove-last"

func (a *App) sendMessage(msg sendMessageRequest) (int64, error) {
	url := fmt.Sprintf("%s%s/sendMessage", TELEGRAM_BOT_BASE_URL, a.config.Token)
	result, err := postTelegramJSON[message](a.httpClient, url, msg)
	if err != nil {
		return 0, err
	}

	return result.ID, nil
}

func (a *App) copyMessage(msg copyMessageRequest) (int64, error) {
	url := fmt.Sprintf("%s%s/copyMessage", TELEGRAM_BOT_BASE_URL, a.config.Token)
	result, err := postTelegramJSON[struct {
		MessageID int64 `json:"message_id"`
	}](a.httpClient, url, msg)

	if err != nil {
		return 0, nil
	}

	return result.MessageID, nil
}

func (a *App) pinMessage(req pinMessageRequest) error {
	url := fmt.Sprintf("%s%s/pinChatMessage", TELEGRAM_BOT_BASE_URL, a.config.Token)

	_, err := postTelegramJSON[any](a.httpClient, url, req)

	return err
}

func (a *App) deleteLastMessage(req deleteMessageRequest) error {
	url := fmt.Sprintf("%s%s/deleteMessage", TELEGRAM_BOT_BASE_URL, a.config.Token)
	_, err := postTelegramJSON[any](a.httpClient, url, req)

	return err
}

func (a *App) sendPhoto(req sendPhotoRequest) (int64, error) {
	url := fmt.Sprintf("%s%s/sendPhoto", TELEGRAM_BOT_BASE_URL, a.config.Token)
	result, err := postTelegramJSON[message](a.httpClient, url, req)
	if err != nil {
		return 0, err
	}

	return result.ID, nil
}

func (b *App) сontrolPanel(chatId int64) error {
	markup := sendMessageRequest{
		ChatID: chatId,
		Text:   "Выберите действие",
		ReplyMarkup: struct {
			InlineKeyboard [][]inlineKeyboardMarkup `json:"inline_keyboard,omitempty"`
		}{
			InlineKeyboard: [][]inlineKeyboardMarkup{
				{
					{
						Text:         "Старт",
						CallbackData: START_CALLBACK_DATA,
					},
				},
				{
					{
						Text:         "Стоп",
						CallbackData: STOP_CALLBACK_DATA,
					},
				},
				{
					{
						Text:         "Добавить чат",
						CallbackData: ADD_CHAT_DATA,
					},
				},
				{
					{
						Text:         "Перезаписать чаты",
						CallbackData: RESET_CHATS_DATA,
					},
				},
				{
					{
						Text:         "Выбрать интервал",
						CallbackData: CHOOSE_INTERVAL_DATA,
					},
				},
				{
					{
						Text:         "Поменять сообщение",
						CallbackData: CHANGE_MESSAGE,
					},
				},
				{
					{
						Text:         "PIN",
						CallbackData: PIN_DATA,
					},
				},
				{
					{
						Text:         "Удалять последние сообщения",
						CallbackData: REMOVE_LAST_DATA,
					},
				},
			},
		},
	}

	_, err := b.sendMessage(markup)

	return err
}

func (a *App) getUpdates(offset int) ([]update, error) {
	url := fmt.Sprintf("%s%s/getUpdates?timeout=30&offset=%d", TELEGRAM_BOT_BASE_URL, a.config.Token, offset)

	resp, err := a.httpClient.Get(url)
	if err != nil {
		if ne, ok := err.(net.Error); ok && ne.Timeout() {
			return nil, nil
		}

		return nil, err
	}
	defer resp.Body.Close()

	var result baseResponse[[]update]

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Ok {
		if result.Description != nil {
			return nil, fmt.Errorf("telegram error: %s", *result.Description)
		}
		return nil, fmt.Errorf("telegram error")
	}

	return result.Result, nil
}

func (a *App) answerCallback(answer callbackAnwser) error {
	url := fmt.Sprintf("%s%s/answerCallbackQuery", TELEGRAM_BOT_BASE_URL, a.config.Token)
	_, err := postTelegramJSON[any](a.httpClient, url, answer)

	return err
}

func decodeTelegramResponse[T any](respBody []byte) (*T, error) {
	var result baseResponse[T]
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Ok {
		if result.Description != nil {
			return nil, fmt.Errorf("telegram error: %s", *result.Description)
		}
		return nil, fmt.Errorf("telegram error")
	}

	return &result.Result, nil
}

func postTelegramJSON[T any](httpClient *http.Client, url string, body any) (*T, error) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return decodeTelegramResponse[T](respBytes)
}
