package config

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
)

type Config struct {
	AdminID     int64   `json:"adminId"`
	Token       string  `json:"-"`
	PostMinute  int64   `json:"postMinute"`
	Pin         bool    `json:"pin"`
	RemoveLast  bool    `json:"removeLast"`
	ChatIDs     []int64 `json:"chatIds"`
	Message     string  `json:"message"`
	PhotoFileID string  `json:"photoFileId,omitempty"`

	path string
}

func New(path string) *Config {
	file, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("failed to read config file: %v", err))
	}

	var cfg Config
	err = json.Unmarshal(file, &cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to parse config JSON: %v", err))
	}

	if cfg.AdminID <= 0 {
		panic("invalid config: adminId must be greater than 0")
	}

	if cfg.PostMinute <= 0 {
		panic("invalid config: postMinute must be greater than 0")
	}

	var token = os.Getenv("BOT_TOKEN")
	if len(strings.TrimSpace(token)) == 0 {
		panic("invalid config: BOT_TOKEN must be set")
	}

	cfg.path = path
	cfg.Token = token

	return &cfg
}

func (c *Config) Save() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(c.path, data, 0644)
}

func (c *Config) AddChat(chatID int64) error {
	if slices.Contains(c.ChatIDs, chatID) {
		return nil
	}

	c.ChatIDs = append(c.ChatIDs, chatID)
	return c.Save()
}

func (c *Config) ResetChats(chatIDs []int64) error {
	c.ChatIDs = chatIDs
	return c.Save()
}

func (c *Config) ChangePostMinute(minutes int64) error {
	if minutes <= 0 {
		return fmt.Errorf("interval must be greater than 0")
	}

	c.PostMinute = minutes
	return c.Save()
}

func (c *Config) ChangeMessage(message, photoFileID string) error {
	if len(strings.TrimSpace(message)) == 0 {
		return fmt.Errorf("message can not be empty")
	}

	c.Message = message
	c.PhotoFileID = photoFileID

	return c.Save()
}

func (c *Config) TogglePin() error {
	c.Pin = !c.Pin

	return c.Save()
}

func (c *Config) ToggleRemoveLast() error {
	c.RemoveLast = !c.RemoveLast

	return c.Save()
}
