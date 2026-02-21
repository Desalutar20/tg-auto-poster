package app

import (
	"context"
	"sync"
	"time"
)

func (a *App) startScheduler(ctx context.Context) {
	ticker := time.NewTicker(time.Minute * time.Duration(a.config.PostMinute))
	defer ticker.Stop()

	a.sendMessages()

	for {
		select {
		case <-ticker.C:
			a.sendMessages()
		case <-ctx.Done():
			a.logger.Info("scheduler stopped")
			return
		}
	}
}

func (a *App) sendMessages() {
	batchSize := 10

	for i := 0; i < len(a.config.ChatIDs); i += batchSize {
		end := min(i+batchSize, len(a.config.ChatIDs))

		var wg sync.WaitGroup

		for _, chatID := range a.config.ChatIDs[i:end] {
			wg.Go(func() {
				a.sendToChat(chatID)
			})
		}

		wg.Wait()

		time.Sleep(2 * time.Second)
	}
}

func (a *App) sendToChat(chatID int64) {
	a.mu.Lock()
	messageId, exists := a.lastMessage[chatID]
	a.mu.Unlock()

	if exists && a.config.RemoveLast {
		if err := a.deleteLastMessage(deleteMessageRequest{
			ChatID:    chatID,
			MessageID: messageId,
		}); err != nil {
			a.logger.Warn("failed to remove last message",
				"chat_id", chatID,
				"error", err,
			)
		}
	}

	var (
		msgID int64
		err   error
	)

	if a.config.PhotoFileID != "" {
		msgID, err = a.sendPhoto(sendPhotoRequest{
			ChatID:    chatID,
			Photo:     a.config.PhotoFileID,
			Caption:   a.config.Message,
			ParseMode: "HTML",
		})
	} else {
		msgID, err = a.sendMessage(sendMessageRequest{
			ChatID:    chatID,
			Text:      a.config.Message,
			ParseMode: "HTML",
		})
	}

	if err != nil {
		a.logger.Error("failed to send message",
			"chat_id", chatID,
			"error", err,
		)
		return
	}

	a.mu.Lock()
	a.lastMessage[chatID] = msgID
	a.mu.Unlock()

	if a.config.Pin {
		if err := a.pinMessage(pinMessageRequest{
			ChatID:    chatID,
			MessageID: msgID,
		}); err != nil {
			a.logger.Warn("failed to pin message",
				"chat_id", chatID,
				"error", err,
			)
		}
	}
}
