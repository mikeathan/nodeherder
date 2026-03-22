import { get, post } from '@/contracts/api';
import type { ConversationHistoryResponse, ConversationsListResponse } from '@/types/assistant.type';
import { fetchWithAuth } from '@/contracts/api';

const CONTEXT_VERSION = 'v1';

export async function sendMessageToLLM(conversationId: string, message: string): Promise<string> {
  const response = await post('assistant/message', {
    conversation_id: conversationId,
    context_version: CONTEXT_VERSION,
    message: message,
  });

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }

  const data = await response.json();
  return data.reply;
}

export async function fetchConversations(): Promise<ConversationsListResponse> {
  const response = await get('assistant/conversations');

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }

  return response.json();
}

export async function fetchConversationHistory(conversationId: string): Promise<ConversationHistoryResponse> {
  const response = await get(`assistant/history/${conversationId}`);

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }

  return response.json();
}

export async function deleteConversation(conversationId: string): Promise<void> {
  const response = await fetchWithAuth(`assistant/history/${conversationId}`, {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
}
