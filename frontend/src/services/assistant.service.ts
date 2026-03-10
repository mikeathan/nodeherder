import type { SendMessagePayload } from '@/types/assistant.type';

const CONTEXT_VERSION = 'v1';

export async function sendMessageToLLM(url: string, conversationId: string, message: string): Promise<string> {
  const response = await fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      conversation_id: conversationId,
      context_version: CONTEXT_VERSION,
      message: message,
    }),
  });

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }

  return await response.text();
}
