package assistant

import (
	"errors"
	"fmt"
	"time"
)

var ErrNotFound = errors.New("conversation not found")

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

type ChatMessage struct {
	ID        string `json:"id"`
	Role      Role   `json:"role"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

type Conversation struct {
	ID        string        `json:"conversation_id"`
	Title     string        `json:"title"`
	Messages  []ChatMessage `json:"messages"`
	CreatedAt time.Time     `json:"created_at"`
}

type ConversationSummary struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	CreatedAt    time.Time `json:"created_at"`
	MessageCount int       `json:"message_count"`
}

const maxTitleLength = 100
const initialMessagesCapacity = 2

func NewConversation(id string, firstMessage string, now time.Time) *Conversation {
	title := firstMessage
	if len(title) > maxTitleLength {
		title = title[:maxTitleLength] + "..."
	}

	return &Conversation{
		ID:        id,
		Title:     title,
		Messages:  make([]ChatMessage, 0, initialMessagesCapacity),
		CreatedAt: now,
	}
}

func (c *Conversation) Summary() ConversationSummary {
	return ConversationSummary{
		ID:           c.ID,
		Title:        c.Title,
		CreatedAt:    c.CreatedAt,
		MessageCount: len(c.Messages),
	}
}

func NewChatMessage(role Role, content string, now time.Time) ChatMessage {
	ts := now.UnixMilli()
	return ChatMessage{
		ID:        fmt.Sprintf("%c-%d", role[0], ts),
		Role:      role,
		Content:   content,
		Timestamp: ts,
	}
}
