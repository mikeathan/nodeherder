package assistant_test

import (
	"node-herder/models/assistant"
	"testing"
	"time"
)

func TestNewConversation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		id            string
		firstMessage  string
		wantTitleLen  int
		wantTruncated bool
	}{
		{
			name:          "short title is not truncated",
			id:            "conv-1",
			firstMessage:  "What's the temperature?",
			wantTitleLen:  len("What's the temperature?"),
			wantTruncated: false,
		},
		{
			name:          "long title is truncated at 100 chars with ellipsis",
			id:            "conv-2",
			firstMessage:  string(make([]byte, 150)),
			wantTitleLen:  103, // 100 + "..."
			wantTruncated: true,
		},
		{
			name:          "exactly 100 chars is not truncated",
			id:            "conv-3",
			firstMessage:  string(make([]byte, 100)),
			wantTitleLen:  100,
			wantTruncated: false,
		},
		{
			name:          "empty message",
			id:            "conv-4",
			firstMessage:  "",
			wantTitleLen:  0,
			wantTruncated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			now := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)
			conv := assistant.NewConversation(tt.id, tt.firstMessage, now)

			if conv.ID != tt.id {
				t.Errorf("ID = %q, want %q", conv.ID, tt.id)
			}

			if len(conv.Title) != tt.wantTitleLen {
				t.Errorf("Title length = %d, want %d", len(conv.Title), tt.wantTitleLen)
			}

			if tt.wantTruncated {
				suffix := conv.Title[len(conv.Title)-3:]
				if suffix != "..." {
					t.Errorf("truncated title should end with '...', got suffix %q", suffix)
				}
			}

			if conv.CreatedAt != now {
				t.Errorf("CreatedAt = %v, want %v", conv.CreatedAt, now)
			}

			if len(conv.Messages) != 0 {
				t.Errorf("Messages should be empty, got %d", len(conv.Messages))
			}
		})
	}
}

func TestConversation_Summary(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)
	conv := &assistant.Conversation{
		ID:    "conv-1",
		Title: "Test conversation",
		Messages: []assistant.ChatMessage{
			{ID: "msg-1", Role: "user", Content: "Hello", Timestamp: 1000},
			{ID: "msg-2", Role: "assistant", Content: "Hi", Timestamp: 2000},
		},
		CreatedAt: now,
	}

	summary := conv.Summary()

	if summary.ID != conv.ID {
		t.Errorf("Summary.ID = %q, want %q", summary.ID, conv.ID)
	}
	if summary.Title != conv.Title {
		t.Errorf("Summary.Title = %q, want %q", summary.Title, conv.Title)
	}
	if summary.MessageCount != 2 {
		t.Errorf("Summary.MessageCount = %d, want 2", summary.MessageCount)
	}
	if summary.CreatedAt != now {
		t.Errorf("Summary.CreatedAt = %v, want %v", summary.CreatedAt, now)
	}
}
