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
