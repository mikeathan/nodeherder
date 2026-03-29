export type Role = 'user' | 'assistant';

export interface AssistantMessage {
  id: string;
  role: Role;
  content: string;
  timestamp: number;
}
export interface SendMessagePayload {
  conversation_id: string;
  context_version: string;
  message: string;
}

export interface Conversation {
  id: string;
  messages: AssistantMessage[];
}

export interface ConversationSummary {
  id: string;
  title: string;
  created_at: string;
  message_count: number;
}

export interface ConversationHistoryResponse {
  conversation_id: string;
  title: string;
  messages: AssistantMessage[];
  created_at: string;
}

export interface ConversationsListResponse {
  conversations: ConversationSummary[];
}
