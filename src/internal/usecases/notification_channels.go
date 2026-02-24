package usecases

import (
	"bytes"
	"context"
	"encoding/json"
	"etruscan/internal/domain/models"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

// TelegramChannel sends notifications to Telegram chats via Bot API.
type TelegramChannel struct {
	botToken string
	logger   *zap.Logger
	client   *http.Client
}

func NewTelegramChannel(botToken string, logger *zap.Logger, client *http.Client) *TelegramChannel {
	return &TelegramChannel{
		botToken: botToken,
		logger:   logger,
		client:   client,
	}
}

func (c *TelegramChannel) Send(ctx context.Context, user *models.User, n models.Notification) error {
	if user.TelegramChatID == nil || *user.TelegramChatID == "" {
		return nil
	}

	text := fmt.Sprintf("*%s*\n\n%s", n.Title, n.Body)

	payload := map[string]any{
		"chat_id":    *user.TelegramChatID,
		"text":       text,
		"parse_mode": "Markdown",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", c.botToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.logger.Warn("telegram sendMessage returned non-2xx",
			zap.Int("status", resp.StatusCode),
		)
	}

	return nil
}

// EmailChannel is an MVP implementation that logs email notifications (can be extended with an SMTP implementation).
type EmailChannel struct {
	from   string
	logger *zap.Logger
}

func NewEmailChannel(from string, logger *zap.Logger) *EmailChannel {
	return &EmailChannel{
		from:   from,
		logger: logger,
	}
}

func (c *EmailChannel) Send(_ context.Context, user *models.User, n models.Notification) error {
	if user.Email == "" {
		return nil
	}

	subject := fmt.Sprintf("[%s] %s", n.Severity, n.Title)

	c.logger.Info("email notification (log sink)",
		zap.String("to", user.Email),
		zap.String("from", c.from),
		zap.String("subject", subject),
		zap.String("body", n.Body),
		zap.Any("metadata", n.Metadata),
	)

	return nil
}
