import { Conversation, ConversationSummary } from '@/types/assistant.type';

export interface AssistantState {
  activeConversation: Conversation;
  conversations: ConversationSummary[];
}
