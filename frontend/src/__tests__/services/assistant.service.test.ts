import { sendMessageToLLM, fetchConversationHistory, fetchConversations, deleteConversation } from '@/services/assistant.service';
const previous = globalThis.fetch;
afterEach(() => { globalThis.fetch = previous; });
test('assistant requests use the shared authenticated helper and forward cancellation signals', async () => {
  const request = jest.fn()
    .mockResolvedValueOnce(new Response(JSON.stringify({ reply: 'answer' })))
    .mockResolvedValueOnce(new Response(JSON.stringify({ conversations: [] })))
    .mockResolvedValueOnce(new Response(JSON.stringify({ conversation_id: 'chat/one', messages: [], title: '', created_at: '' })))
    .mockResolvedValueOnce(new Response(null, { status: 204 }));
  globalThis.fetch = request;
  const controller = new AbortController();
  expect(await sendMessageToLLM('chat', 'question', controller.signal)).toBe('answer');
  await fetchConversations(controller.signal);
  await fetchConversationHistory('chat/one', controller.signal);
  await deleteConversation('chat/one', controller.signal);
  expect(request).toHaveBeenCalledTimes(4);
  for (const [, options] of request.mock.calls) {
    expect(options.signal).toBe(controller.signal);
    expect(options.credentials).toBe('include');
  }
  expect(request.mock.calls[2][0]).toContain('/assistant/history/chat%2Fone');
});
