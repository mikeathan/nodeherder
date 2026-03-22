package store

import (
	"errors"
	"fmt"
	"node-herder/models/assistant"
	"time"
)

func (s *appStore) AppendAssistantMessage(conversationID string, role assistant.Role, content string) error {
	if s.assistantRepo == nil {
		return fmt.Errorf("assistant repository not initialized")
	}
	if conversationID == "" {
		return fmt.Errorf("conversation ID cannot be empty")
	}
	if role == "" {
		return fmt.Errorf("role cannot be empty")
	}
	if content == "" {
		return fmt.Errorf("message content cannot be empty")
	}

	now := time.Now()
	conv, err := s.assistantRepo.LoadConversation(conversationID)
	if err != nil {
		if !errors.Is(err, assistant.ErrNotFound) {
			return err
		}
		conv = assistant.NewConversation(conversationID, content, now)
	}

	msg := assistant.NewChatMessage(role, content, now)

	conv.Messages = append(conv.Messages, msg)
	return s.assistantRepo.SaveConversation(conv)
}

func (s *appStore) LoadAssistantHistory(conversationID string) (*assistant.Conversation, error) {
	if s.assistantRepo == nil {
		return nil, fmt.Errorf("assistant repository not initialized")
	}
	return s.assistantRepo.LoadConversation(conversationID)
}

func (s *appStore) ListAssistantConversations() ([]assistant.ConversationSummary, error) {
	if s.assistantRepo == nil {
		return nil, fmt.Errorf("assistant repository not initialized")
	}
	return s.assistantRepo.ListConversations()
}

func (s *appStore) DeleteAssistantConversation(conversationID string) error {
	if s.assistantRepo == nil {
		return fmt.Errorf("assistant repository not initialized")
	}
	return s.assistantRepo.DeleteConversation(conversationID)
}
