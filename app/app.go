package app

import (
	"context"
	"fmt"
	"go-bot/config"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type App struct {
	config                 *config.Config
	httpClient             *http.Client
	logger                 *slog.Logger
	schedulerCtx           context.Context
	schedulerCtxCancelFunc context.CancelFunc
	lastMessage            map[int64]int64
	mu                     sync.Mutex
	callbackType           Callback
}

func (a *App) Run(ctx context.Context) {
	a.schedulerCtx, a.schedulerCtxCancelFunc = context.WithCancel(ctx)
	go a.startScheduler(a.schedulerCtx)

	var offset int

	for {
		select {
		case <-ctx.Done():

			a.httpClient.CloseIdleConnections()
			return
		default:

			updates, err := a.getUpdates(offset)
			if err != nil {
				a.logger.Warn("failed to get updates", "error", err)
				time.Sleep(5 * time.Second)

				continue
			}

			for _, u := range updates {
				offset = u.ID + 1

				if u.Message != nil {
					if u.Message.From == nil || u.Message.From.ID != a.config.AdminID {
						continue
					}

					a.handleMessage(u.Message, ctx)
				}

				if u.CallbackQuery != nil {
					if u.CallbackQuery.From.ID != a.config.AdminID {
						continue
					}

					a.handleCallback(u.CallbackQuery, ctx)
				}
			}

			time.Sleep(2 * time.Second)
		}
	}
}

func (a *App) handleCallback(cb *callbackQuery, ctx context.Context) {
	answer := callbackAnwser{
		ID: cb.ID,
	}

	callbackType := Callback(cb.Data)
	callPanel := false

	switch callbackType {
	case START_CALLBACK_DATA:
		answer.ShowAlert = true
		callPanel = true

		if a.schedulerCtx != nil {
			a.logger.Info("Scheduler already running")
			answer.Text = "Автопостинг уже запущен ✅"

			break
		}

		a.logger.Info("Starting scheduler")

		a.schedulerCtx, a.schedulerCtxCancelFunc = context.WithCancel(ctx)
		go a.startScheduler(a.schedulerCtx)

		answer.Text = "Автопостинг запущен ✅"

	case STOP_CALLBACK_DATA:
		{
			answer.ShowAlert = true
			callPanel = true

			if a.schedulerCtxCancelFunc != nil {
				a.logger.Info("Stopping scheduler")

				a.schedulerCtxCancelFunc()
				a.schedulerCtx = nil
				a.schedulerCtxCancelFunc = nil

				answer.Text = "Автопостинг остановлен ⏹"

				break
			}

			a.logger.Info("Scheduler is not running")
			answer.Text = "Автопостинг ещё не запущен ⚠️"
		}

	case ADD_CHAT_DATA:
		{
			if _, err := a.sendMessage(sendMessageRequest{
				ChatID: cb.Message.Chat.ID,
				Text:   "Введите id чата",
			}); err != nil {
				a.logger.Warn(err.Error())
				break
			}

		}

	case RESET_CHATS_DATA:
		{
			if _, err := a.sendMessage(sendMessageRequest{
				ChatID: cb.Message.Chat.ID,
				Text:   "Введите id чатов разделенный пробелами",
			}); err != nil {
				a.logger.Warn(err.Error())
				break
			}
		}

	case CHOOSE_INTERVAL_DATA:
		{

			if _, err := a.sendMessage(sendMessageRequest{
				ChatID: cb.Message.Chat.ID,
				Text:   "Введите интервал в минутах",
			}); err != nil {
				a.logger.Warn(err.Error())
				break
			}
		}

	case CHANGE_MESSAGE:
		{

			if _, err := a.sendMessage(sendMessageRequest{
				ChatID: cb.Message.Chat.ID,
				Text:   "Введите новое сообщение",
			}); err != nil {
				a.logger.Warn(err.Error())
				break
			}

		}

	case PIN_DATA:
		{
			answer.ShowAlert = true

			if err := a.config.TogglePin(); err != nil {
				a.logger.Warn(err.Error())
				answer.Text = "❌ Не удалось изменить состояние закрепления"
				break
			}

			if a.config.Pin {
				answer.Text = "📌 Сообщения теперь будут закрепляться"
			} else {
				answer.Text = "📍 Сообщения больше не будут закрепляться"
			}

			callPanel = true
		}

	case REMOVE_LAST_DATA:
		{
			answer.ShowAlert = true

			if err := a.config.ToggleRemoveLast(); err != nil {
				a.logger.Warn(err.Error())
				answer.Text = "❌ Не удалось изменить состояние удаления"
				break
			}

			if a.config.RemoveLast {
				answer.Text = "🗑 Сообщения теперь будут удаляться перед отправкой новых"
			} else {
				answer.Text = "✅ Сообщения больше не будут удаляться автоматически"
			}

			callPanel = true
		}

	}

	a.callbackType = callbackType
	a.answerCallback(answer)

	if callPanel {
		a.сontrolPanel(cb.Message.Chat.ID)
	}
}

func (a *App) handleMessage(msg *message, ctx context.Context) {
	message := msg.Text

	if msg.Text == "/start" {
		if err := a.сontrolPanel(msg.Chat.ID); err != nil {
			a.logger.Warn("failed to send control panel", "chat_id", msg.Chat.ID, "error", err)
		}

		return
	}

	if a.callbackType == ADD_CHAT_DATA {
		chatID, err := strconv.ParseInt(message, 10, 64)
		if err != nil {
			if _, sendErr := a.sendMessage(sendMessageRequest{
				ChatID: msg.Chat.ID,
				Text:   "❌ Некорректный ID чата. Введите числовой ID чата:",
			}); sendErr != nil {
				a.logger.Warn(sendErr.Error())
			}

			a.сontrolPanel(msg.Chat.ID)
			return
		}

		if err := a.config.AddChat(chatID); err != nil {
			a.logger.Warn("failed to add chat", "chat_id", chatID, "error", err)
			return
		}

		if _, sendErr := a.sendMessage(sendMessageRequest{
			ChatID: msg.Chat.ID,
			Text:   fmt.Sprintf("✅ Чат %d успешно добавлен", chatID),
		}); sendErr != nil {
			a.logger.Warn(sendErr.Error())
		}

		a.callbackType = NONE_DATA
		a.сontrolPanel(msg.Chat.ID)

		return
	}

	if a.callbackType == RESET_CHATS_DATA {
		parts := strings.Fields(message)
		if len(parts) == 0 {
			return
		}

		var parsedIDs []int64
		for _, s := range parts {
			id, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				continue
			}

			parsedIDs = append(parsedIDs, id)
		}

		if len(parsedIDs) == 0 {
			return
		}

		if err := a.config.ResetChats(parsedIDs); err != nil {
			a.logger.Warn("failed to reset chats", "error", err)
			return
		}

		if _, sendErr := a.sendMessage(sendMessageRequest{
			ChatID: msg.Chat.ID,
			Text:   fmt.Sprintf("✅ Список чатов успешно перезаписан: %v", parsedIDs),
		}); sendErr != nil {
			a.logger.Warn(sendErr.Error())
		}

		a.callbackType = NONE_DATA
		a.сontrolPanel(msg.Chat.ID)

		return
	}

	if a.callbackType == CHOOSE_INTERVAL_DATA {
		parsed, err := strconv.ParseInt(message, 10, 64)
		if err != nil || parsed <= 0 {
			if _, sendErr := a.sendMessage(sendMessageRequest{
				ChatID: msg.Chat.ID,
				Text:   "❌ Некорректный интервал. Введите числовое значение (> 0)",
			}); sendErr != nil {
				a.logger.Warn(sendErr.Error())
			}

			return
		}

		if err := a.config.ChangePostMinute(parsed); err != nil {
			a.logger.Warn("failed to change post interval", "error", err)
			return
		}

		if a.schedulerCtxCancelFunc != nil {
			a.schedulerCtxCancelFunc()

			a.schedulerCtx, a.schedulerCtxCancelFunc = context.WithCancel(ctx)
			go a.startScheduler(a.schedulerCtx)
		}

		if _, sendErr := a.sendMessage(sendMessageRequest{
			ChatID: msg.Chat.ID,
			Text:   "✅ Интервал автопостинга успешно изменен",
		}); sendErr != nil {
			a.logger.Warn(sendErr.Error())
		}

		a.callbackType = NONE_DATA
		a.сontrolPanel(msg.Chat.ID)

		return
	}

	if a.callbackType == CHANGE_MESSAGE {
		var text string
		var entities []messageEntity
		var photoFileID string

		if msg.Caption != nil {
			text = *msg.Caption
			entities = msg.CaptionEntities
		} else {
			text = msg.Text
			entities = msg.Entities
		}

		if len(msg.Photo) > 0 {
			photo := msg.Photo[len(msg.Photo)-1]
			photoFileID = photo.FileID
		}

		if len(msg.CaptionEntities) > 0 {
			entities = msg.CaptionEntities
		} else {
			entities = msg.Entities
		}

		if len(strings.TrimSpace(text)) == 0 {
			if _, sendErr := a.sendMessage(sendMessageRequest{
				ChatID: msg.Chat.ID,
				Text:   "Сообщение не может быть пустым",
			}); sendErr != nil {
				a.logger.Warn(sendErr.Error())
			}
			return
		}

		if len(entities) > 0 {
			text = UnparseEntitiesToHTML(text, entities)
		}

		if err := a.config.ChangeMessage(text, photoFileID); err != nil {
			a.logger.Warn("failed to change message", "error", err)
			return
		}

		if _, sendErr := a.sendMessage(sendMessageRequest{
			ChatID: msg.Chat.ID,
			Text:   "✅ Сообщение успешно изменен",
		}); sendErr != nil {
			a.logger.Warn(sendErr.Error())
		}

		a.callbackType = NONE_DATA
		a.сontrolPanel(msg.Chat.ID)

		return
	}

}

func New(cfg *config.Config, logger *slog.Logger) *App {
	return &App{
		config:      cfg,
		httpClient:  &http.Client{Timeout: 35 * time.Second},
		logger:      logger,
		lastMessage: make(map[int64]int64),
	}
}
