import { AssistantMessage, Role, ConversationHistoryResponse, ConversationsListResponse } from '@/types/assistant.type';
import { genenerateUniqueId } from '@/utils/unique';

export const createUserRequest = (message: string) => {
  return createMessage('user', message);
};

export const createAssistantResponse = (message: string) => {
  return createMessage('assistant', message);
};

export const createAssistantErrorResponse = (error: Error) => {
  return createMessage('assistant', `Error: ${error.message}`);
};

const createMessage = (role: Role, message: string) :AssistantMessage=> {
  return {
    id: genenerateUniqueId(),
    role,
    content: message,
    timestamp: Date.now(),
  };
};


function record(value: unknown): Record<string, unknown> {
  if (typeof value !== 'object' || value === null || Array.isArray(value)) throw new Error('Invalid assistant response');
  return value as Record<string, unknown>;
}
function text(value: unknown): string {
  if (typeof value !== 'string') throw new Error('Invalid assistant text');
  return value;
}
function message(value: unknown): AssistantMessage {
  const data = record(value);
  if ((data.role !== 'user' && data.role !== 'assistant') || typeof data.timestamp !== 'number' || !Number.isFinite(data.timestamp)) {
    throw new Error('Invalid assistant message');
  }
  return { id: text(data.id), role: data.role, content: text(data.content), timestamp: data.timestamp };
}
export function parseAssistantReply(value: unknown): string {
  return text(record(value).reply);
}
export function parseConversationHistory(value: unknown, expectedId: string): ConversationHistoryResponse {
  const data = record(value);
  if (data.conversation_id !== expectedId) throw new Error('Unexpected assistant conversation');
  const messages = data.messages ?? [];
  if (!Array.isArray(messages)) throw new Error('Invalid assistant history');
  return { conversation_id: expectedId, messages: messages.map(message), title: text(data.title), created_at: text(data.created_at) };
}
export function parseConversations(value: unknown): ConversationsListResponse {
  const conversations = record(value).conversations ?? [];
  if (!Array.isArray(conversations)) throw new Error('Invalid assistant conversations');
  return { conversations: conversations.map((value) => {
    const data = record(value);
    if (typeof data.message_count !== 'number' || !Number.isSafeInteger(data.message_count) || data.message_count < 0) throw new Error('Invalid assistant message count');
    return { id: text(data.id), title: text(data.title), created_at: text(data.created_at), message_count: data.message_count };
  }) };
}
