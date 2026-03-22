package assistant

type Repository interface {
	SaveConversation(conv *Conversation) error
	LoadConversation(id string) (*Conversation, error)
	ListConversations() ([]ConversationSummary, error)
	DeleteConversation(id string) error
	Close() error
}
