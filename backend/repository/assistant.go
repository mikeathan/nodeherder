package repository

import (
	"encoding/json"
	"fmt"
	"node-herder/models/assistant"
	"node-herder/utils"
	"node-herder/utils/storage"
	"path/filepath"
	"sort"
)

const assistantBaseFilename = "assistant.db"
const assistantBucketName = "assistant_chats"

type AssistantRepo struct {
	kvdb storage.KeyValueDatabase
}

func NewAssistantRepo() (assistant.Repository, error) {
	return NewAssistantRepoFromFile(filepath.Join(utils.GetDataDir(), assistantBaseFilename))
}

func NewAssistantRepoFromFile(filename string) (assistant.Repository, error) {
	kvdb, err := storage.NewBoltKeyValueDatabase(filename, assistantBucketName)
	if err != nil {
		return nil, err
	}

	return &AssistantRepo{kvdb: kvdb}, nil
}

func (r *AssistantRepo) SaveConversation(conv *assistant.Conversation) error {
	if conv.ID == "" {
		return fmt.Errorf("conversation ID cannot be empty")
	}

	buf, err := json.Marshal(conv)
	if err != nil {
		return fmt.Errorf("failed to marshal conversation: %w", err)
	}

	return r.kvdb.Set([]byte(conv.ID), buf)
}

func (r *AssistantRepo) LoadConversation(id string) (*assistant.Conversation, error) {
	buf, err := r.kvdb.Get([]byte(id))
	if err != nil {
		return nil, fmt.Errorf("failed to load conversation: %w", err)
	}

	if buf == nil {
		return nil, assistant.ErrNotFound
	}

	conv := &assistant.Conversation{}
	if err := json.Unmarshal(buf, conv); err != nil {
		return nil, fmt.Errorf("failed to unmarshal conversation: %w", err)
	}

	return conv, nil
}

func (r *AssistantRepo) ListConversations() ([]assistant.ConversationSummary, error) {
	summaries := make([]assistant.ConversationSummary, 0)

	err := r.kvdb.GetAll(func(key, value []byte) error {
		conv := &assistant.Conversation{}
		if err := json.Unmarshal(value, conv); err != nil {
			return nil
		}
		summaries = append(summaries, conv.Summary())
		return nil
	})
	if err != nil {
		return summaries, nil
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].CreatedAt.After(summaries[j].CreatedAt)
	})

	return summaries, nil
}

func (r *AssistantRepo) DeleteConversation(id string) error {
	return r.kvdb.DeleteKey([]byte(id))
}

func (r *AssistantRepo) Close() error {
	return r.kvdb.Close()
}
