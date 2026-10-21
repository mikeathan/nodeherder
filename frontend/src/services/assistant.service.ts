import { get, post } from '@/contracts/api';
import { parseAssistantReply, parseConversationHistory, parseConversations } from '@/contracts/assistant';
import type { ConversationHistoryResponse, ConversationsListResponse } from '@/types/assistant.type';
import { fetchWithAuth } from '@/contracts/api';

const CONTEXT_VERSION = 'v1';

export async function sendMessageToLLM(conversationId: string, message: string, signal?: AbortSignal): Promise<string> {
  const response = await post('assistant/message', {
    conversation_id: conversationId,
    context_version: CONTEXT_VERSION,
    message: message,
  }, { signal });

  if (!response.ok) {
    let errorMessage = `HTTP error! status: ${response.status}`;
    try {
      const errorData = await response.json();
      errorMessage = errorData.error || errorData.message || errorData.status || JSON.stringify(errorData);
    } catch (_) {}
    throw new Error(errorMessage);
  }

  return parseAssistantReply(await response.json());
}

export async function fetchConversations(signal?: AbortSignal): Promise<ConversationsListResponse> {
  const response = await get('assistant/conversations', { signal });

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }

  return parseConversations(await response.json());
}

export async function fetchConversationHistory(conversationId: string, signal?: AbortSignal): Promise<ConversationHistoryResponse> {
  const response = await get(`assistant/history/${encodeURIComponent(conversationId)}`, { signal });

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }

  return parseConversationHistory(await response.json(), conversationId);
}

export async function deleteConversation(conversationId: string, signal?: AbortSignal): Promise<void> {
  const response = await fetchWithAuth(`assistant/history/${encodeURIComponent(conversationId)}`, {
    method: 'DELETE',
    signal,
  });

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
}
