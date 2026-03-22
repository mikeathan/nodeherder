package repository_test

import (
	"errors"
	"node-herder/models/assistant"
	"node-herder/repository"
	utils_test "node-herder/testing"
	"os"
	"testing"
	"time"
)

func TestAssistantRepoSaveAndLoad(t *testing.T) {
	tempfile := utils_test.Tempfile()
	defer os.Remove(tempfile)

	repo, err := repository.NewAssistantRepoFromFile(tempfile)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}
	defer repo.Close()

	now := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)
	conv := &assistant.Conversation{
		ID:    "conv-1",
		Title: "Test conversation",
		Messages: []assistant.ChatMessage{
			{ID: "msg-1", Role: "user", Content: "Hello", Timestamp: now.UnixMilli()},
			{ID: "msg-2", Role: "assistant", Content: "Hi there", Timestamp: now.Add(time.Second).UnixMilli()},
		},
		CreatedAt: now,
	}

	if err := repo.SaveConversation(conv); err != nil {
		t.Fatalf("SaveConversation failed: %v", err)
	}

	loaded, err := repo.LoadConversation("conv-1")
	if err != nil {
		t.Fatalf("LoadConversation failed: %v", err)
	}

	if loaded.ID != conv.ID {
		t.Errorf("ID = %q, want %q", loaded.ID, conv.ID)
	}
	if loaded.Title != conv.Title {
		t.Errorf("Title = %q, want %q", loaded.Title, conv.Title)
	}
	if len(loaded.Messages) != 2 {
		t.Fatalf("Messages count = %d, want 2", len(loaded.Messages))
	}
	if loaded.Messages[0].Role != "user" {
		t.Errorf("Messages[0].Role = %q, want 'user'", loaded.Messages[0].Role)
	}
	if loaded.Messages[1].Content != "Hi there" {
		t.Errorf("Messages[1].Content = %q, want 'Hi there'", loaded.Messages[1].Content)
	}
}

func TestAssistantRepoLoadNotFound(t *testing.T) {
	tempfile := utils_test.Tempfile()
	defer os.Remove(tempfile)

	repo, err := repository.NewAssistantRepoFromFile(tempfile)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}
	defer repo.Close()

	_, err = repo.LoadConversation("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent conversation, got nil")
	}
	if !errors.Is(err, assistant.ErrNotFound) {
		t.Errorf("expected assistant.ErrNotFound, got %v", err)
	}
}

func TestAssistantRepoSaveEmptyID(t *testing.T) {
	tempfile := utils_test.Tempfile()
	defer os.Remove(tempfile)

	repo, err := repository.NewAssistantRepoFromFile(tempfile)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}
	defer repo.Close()

	conv := &assistant.Conversation{ID: "", Title: "test"}
	err = repo.SaveConversation(conv)
	if err == nil {
		t.Error("expected error for empty ID, got nil")
	}
}

func TestAssistantRepoListConversations(t *testing.T) {
	tempfile := utils_test.Tempfile()
	defer os.Remove(tempfile)

	repo, err := repository.NewAssistantRepoFromFile(tempfile)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}
	defer repo.Close()

	// Create three conversations with different times
	now := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)
	conversations := []*assistant.Conversation{
		{
			ID:        "conv-oldest",
			Title:     "Oldest",
			Messages:  []assistant.ChatMessage{{ID: "m1", Role: "user", Content: "one", Timestamp: 1}},
			CreatedAt: now.Add(-2 * time.Hour),
		},
		{
			ID:        "conv-newest",
			Title:     "Newest",
			Messages:  []assistant.ChatMessage{{ID: "m2", Role: "user", Content: "two", Timestamp: 2}},
			CreatedAt: now,
		},
		{
			ID:    "conv-middle",
			Title: "Middle",
			Messages: []assistant.ChatMessage{
				{ID: "m3", Role: "user", Content: "three", Timestamp: 3},
				{ID: "m4", Role: "assistant", Content: "four", Timestamp: 4},
			},
			CreatedAt: now.Add(-1 * time.Hour),
		},
	}

	for _, conv := range conversations {
		if err := repo.SaveConversation(conv); err != nil {
			t.Fatalf("SaveConversation failed: %v", err)
		}
	}

	summaries, err := repo.ListConversations()
	if err != nil {
		t.Fatalf("ListConversations failed: %v", err)
	}

	if len(summaries) != 3 {
		t.Fatalf("expected 3 summaries, got %d", len(summaries))
	}

	// Should be sorted newest first
	if summaries[0].ID != "conv-newest" {
		t.Errorf("first summary ID = %q, want 'conv-newest'", summaries[0].ID)
	}
	if summaries[1].ID != "conv-middle" {
		t.Errorf("second summary ID = %q, want 'conv-middle'", summaries[1].ID)
	}
	if summaries[2].ID != "conv-oldest" {
		t.Errorf("third summary ID = %q, want 'conv-oldest'", summaries[2].ID)
	}

	// Verify message counts
	if summaries[1].MessageCount != 2 {
		t.Errorf("middle conversation MessageCount = %d, want 2", summaries[1].MessageCount)
	}
}

func TestAssistantRepoDeleteConversation(t *testing.T) {
	tempfile := utils_test.Tempfile()
	defer os.Remove(tempfile)

	repo, err := repository.NewAssistantRepoFromFile(tempfile)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}
	defer repo.Close()

	conv := &assistant.Conversation{
		ID:        "conv-delete",
		Title:     "To be deleted",
		Messages:  []assistant.ChatMessage{},
		CreatedAt: time.Now(),
	}

	if err := repo.SaveConversation(conv); err != nil {
		t.Fatalf("SaveConversation failed: %v", err)
	}

	// Verify it exists
	_, err = repo.LoadConversation("conv-delete")
	if err != nil {
		t.Fatalf("conversation should exist before delete: %v", err)
	}

	// Delete
	if err := repo.DeleteConversation("conv-delete"); err != nil {
		t.Fatalf("DeleteConversation failed: %v", err)
	}

	// Verify it's gone
	_, err = repo.LoadConversation("conv-delete")
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
	if !errors.Is(err, assistant.ErrNotFound) {
		t.Errorf("expected assistant.ErrNotFound, got %v", err)
	}
}

func TestAssistantRepoListEmpty(t *testing.T) {
	tempfile := utils_test.Tempfile()
	defer os.Remove(tempfile)

	repo, err := repository.NewAssistantRepoFromFile(tempfile)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}
	defer repo.Close()

	summaries, err := repo.ListConversations()
	if err != nil {
		t.Fatalf("ListConversations on empty db failed: %v", err)
	}

	if len(summaries) != 0 {
		t.Errorf("expected 0 summaries, got %d", len(summaries))
	}
}

func TestAssistantRepoUpdateConversation(t *testing.T) {
	tempfile := utils_test.Tempfile()
	defer os.Remove(tempfile)

	repo, err := repository.NewAssistantRepoFromFile(tempfile)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}
	defer repo.Close()

	now := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)
	conv := &assistant.Conversation{
		ID:        "conv-update",
		Title:     "First message",
		Messages:  []assistant.ChatMessage{{ID: "m1", Role: "user", Content: "Hello", Timestamp: now.UnixMilli()}},
		CreatedAt: now,
	}

	if err := repo.SaveConversation(conv); err != nil {
		t.Fatalf("SaveConversation failed: %v", err)
	}

	// Append a message and save again
	conv.Messages = append(conv.Messages, assistant.ChatMessage{
		ID: "m2", Role: "assistant", Content: "Hi!", Timestamp: now.Add(time.Second).UnixMilli(),
	})

	if err := repo.SaveConversation(conv); err != nil {
		t.Fatalf("SaveConversation (update) failed: %v", err)
	}

	loaded, err := repo.LoadConversation("conv-update")
	if err != nil {
		t.Fatalf("LoadConversation failed: %v", err)
	}

	if len(loaded.Messages) != 2 {
		t.Errorf("Messages count = %d, want 2", len(loaded.Messages))
	}
}
