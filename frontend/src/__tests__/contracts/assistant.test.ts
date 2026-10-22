import { parseAssistantReply, parseConversationHistory, parseConversations } from '@/contracts/assistant';
test('assistant contracts preserve plain text and empty history/list shapes', () => {
  expect(parseAssistantReply({ reply: '<script>untrusted text</script>' })).toBe('<script>untrusted text</script>');
  expect(parseConversations({ conversations: null })).toEqual({ conversations: [] });
  expect(parseConversationHistory({ conversation_id: 'c', messages: null, title: '', created_at: '' }, 'c').messages).toEqual([]);
});
test('rejects another conversation, malformed messages, invalid replies and counts', () => {
  expect(() => parseAssistantReply({ reply: {} })).toThrow();
  expect(() => parseConversationHistory({ conversation_id: 'other' }, 'c')).toThrow();
  expect(() => parseConversationHistory({ conversation_id: 'c', messages: [{ role: 'system', timestamp: 0 }] }, 'c')).toThrow();
  expect(() => parseConversations({ conversations: [{ message_count: -1 }] })).toThrow();
});
