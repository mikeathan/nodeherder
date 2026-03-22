package store_test

import (
	"errors"
	"node-herder/mocks"
	"node-herder/models/assistant"
	"node-herder/store"
	"testing"
)

func TestAppendAssistantMessage_NewConversation(t *testing.T) {
	mockRepo := mocks.NewMockAssistantRepo()
	s, _ := store.NewAppStore(nil, nil, nil, mockRepo)

	convID := "new-conv"
	role := assistant.RoleUser
	content := "Hello world"

	mockRepo.OnLoad = func(id string) (*assistant.Conversation, error) {
		return nil, assistant.ErrNotFound
	}

	err := s.AppendAssistantMessage(convID, role, content)
	if err != nil {
		t.Fatalf("AppendAssistantMessage failed: %v", err)
	}

	if !mockRepo.SaveCalled {
		t.Fatal("Expected SaveConversation to be called")
	}

	saved := mockRepo.SavedConversation
	if saved.ID != convID {
		t.Errorf("Saved ID = %q, want %q", saved.ID, convID)
	}
	if len(saved.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(saved.Messages))
	}
	if saved.Messages[0].Content != content {
		t.Errorf("Message content = %q, want %q", saved.Messages[0].Content, content)
	}
	if saved.Messages[0].Role != role {
		t.Errorf("Message role = %q, want %q", saved.Messages[0].Role, role)
	}
}

func TestAppendAssistantMessage_ExistingConversation(t *testing.T) {
	mockRepo := mocks.NewMockAssistantRepo()
	s, _ := store.NewAppStore(nil, nil, nil, mockRepo)

	convID := "existing-conv"
	existingConv := &assistant.Conversation{
		ID:       convID,
		Messages: []assistant.ChatMessage{{ID: "msg-0", Role: assistant.RoleUser, Content: "Old message"}},
	}
	mockRepo.Conversations[convID] = existingConv

	role := assistant.RoleAssistant
	content := "I'm here to help"

	err := s.AppendAssistantMessage(convID, role, content)
	if err != nil {
		t.Fatalf("AppendAssistantMessage failed: %v", err)
	}

	if len(existingConv.Messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(existingConv.Messages))
	}
	if existingConv.Messages[1].Content != content {
		t.Errorf("New message content = %q, want %q", existingConv.Messages[1].Content, content)
	}
	if existingConv.Messages[1].Role != role {
		t.Errorf("New message role = %q, want %q", existingConv.Messages[1].Role, role)
	}
}

func TestAppendAssistantMessage_DatabaseError(t *testing.T) {
	mockRepo := mocks.NewMockAssistantRepo()
	s, _ := store.NewAppStore(nil, nil, nil, mockRepo)

	expectedErr := errors.New("db disconnect")
	mockRepo.OnLoad = func(id string) (*assistant.Conversation, error) {
		return nil, expectedErr
	}

	err := s.AppendAssistantMessage("any-id", assistant.RoleUser, "msg")
	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}

	if mockRepo.SaveCalled {
		t.Fatal("SaveConversation should NOT be called on database load error")
	}
}
