import { effectScope, EffectScope, onMounted } from 'vue';
import { store } from '@/store';
import { createAppconfig } from '@/contracts/settings';
import { useAssistant } from '@/composables/useAssistant';
import { sendMessageToLLM, fetchConversationHistory, fetchConversations, deleteConversation } from '@/services/assistant.service';
import type { ConversationHistoryResponse } from '@/types/assistant.type';

jest.mock('vue', () => ({ ...jest.requireActual('vue'), onMounted: jest.fn() }));
jest.mock('@/services/assistant.service', () => ({
  sendMessageToLLM: jest.fn(), fetchConversationHistory: jest.fn(), fetchConversations: jest.fn(), deleteConversation: jest.fn(),
}));
function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason: Error) => void;
  const promise = new Promise<T>((yes, no) => { resolve = yes; reject = no; });
  return { promise, resolve, reject };
}
function history(id: string): ConversationHistoryResponse {
  return { conversation_id: id, messages: [], title: id, created_at: '' };
}
async function settle() { for (let i = 0; i < 8; i++) await Promise.resolve(); }
let scope: EffectScope;
let assistant: ReturnType<typeof useAssistant>;
beforeEach(() => {
  jest.useFakeTimers(); jest.clearAllMocks();
  const config = createAppconfig(); config.hub.assistant.url = 'http://assistant.invalid'; config.hub.mcp.enabled = true;
  store.commit('hub/setAppConfig', config);
  store.commit('hub/setMCPStatus', { running: true, connectedClients: 1 });
  store.commit('assistant/setActiveConversation', { id: 'original', messages: [] });
  scope = effectScope(); assistant = scope.run(useAssistant)!;
});
afterEach(() => { scope.stop(); expect(jest.getTimerCount()).toBe(0); jest.useRealTimers(); });

test('successful replies and failures stay in their conversation', async () => {
  jest.mocked(sendMessageToLLM).mockResolvedValueOnce('reply');
  assistant.inputText.value = 'question'; await assistant.sendMessage();
  expect(assistant.activeConversation.value.messages.map((m) => m.content)).toEqual(['question', 'reply']);
  jest.mocked(sendMessageToLLM).mockRejectedValueOnce(new Error('provider unavailable'));
  assistant.inputText.value = 'again'; await assistant.sendMessage();
  expect(assistant.activeConversation.value.messages[3].content).toBe('Error: provider unavailable');
  expect(assistant.isLoading.value).toBe(false);
});

test('new conversation aborts pending reply and ignores even an upstream ignoring cancellation', async () => {
  const reply = deferred<string>(); jest.mocked(sendMessageToLLM).mockReturnValue(reply.promise);
  assistant.inputText.value = 'private question'; const pending = assistant.sendMessage(); await settle();
  const args = jest.mocked(sendMessageToLLM).mock.calls[0]; expect(args[0]).toBe('original');
  assistant.startNewConversation(); expect(args[2]?.aborted).toBe(true);
  reply.resolve('late reply'); await pending;
  expect(assistant.activeConversation.value.id).not.toBe('original');
  expect(assistant.activeConversation.value.messages).toEqual([]);
});

test('last history selection wins and a new chat invalidates history', async () => {
  const first = deferred<ConversationHistoryResponse>(), second = deferred<ConversationHistoryResponse>();
  jest.mocked(fetchConversationHistory).mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise);
  const a = assistant.selectConversation('a'), b = assistant.selectConversation('b');
  second.resolve(history('b')); await b; first.resolve(history('a')); await a;
  expect(assistant.activeConversation.value.id).toBe('b');
  const late = deferred<ConversationHistoryResponse>(); jest.mocked(fetchConversationHistory).mockReturnValueOnce(late.promise);
  const pending = assistant.selectConversation('old'); assistant.startNewConversation();
  const id = assistant.activeConversation.value.id;
  late.resolve(history('old')); await pending;
  expect(assistant.activeConversation.value.id).toBe(id);
});

test('logout clears state, aborts pending work and rejects late list/reply data', async () => {
  const reply = deferred<string>(); const list = deferred<Awaited<ReturnType<typeof fetchConversations>>>();
  jest.mocked(sendMessageToLLM).mockReturnValueOnce(reply.promise);
  jest.mocked(fetchConversations).mockReturnValueOnce(list.promise);
  assistant.inputText.value = 'question'; const message = assistant.sendMessage(); const conversations = assistant.loadConversations();
  await settle(); store.commit('auth/logout');
  expect(jest.mocked(sendMessageToLLM).mock.calls[0][2]?.aborted).toBe(true);
  reply.resolve('late private text'); list.resolve({ conversations: [{ id: 'private', title: 'secret', created_at: '', message_count: 1 }] });
  await Promise.all([message, conversations]);
  expect(assistant.activeConversation.value.messages).toEqual([]);
  expect(assistant.conversations.value).toEqual([]);
  expect(assistant.isLoading.value).toBe(false);
});

test('unmount aborts requests and a canceled hydration cannot launch another request', async () => {
  const response = deferred<ConversationHistoryResponse>();
  jest.mocked(fetchConversationHistory).mockReturnValue(response.promise);
  const mounted = jest.mocked(onMounted).mock.calls[0][0] as () => Promise<void>;
  const pending = mounted(); await settle();
  scope.stop(); expect(jest.mocked(fetchConversationHistory).mock.calls[0][1]?.aborted).toBe(true);
  response.resolve(history('original')); await pending;
  expect(fetchConversations).not.toHaveBeenCalled();
});

test('deleting a pending selection cannot resurrect its history', async () => {
  const response = deferred<ConversationHistoryResponse>();
  jest.mocked(fetchConversationHistory).mockReturnValueOnce(response.promise);
  jest.mocked(deleteConversation).mockResolvedValueOnce();
  const pending = assistant.selectConversation('deleted');
  const signal = jest.mocked(fetchConversationHistory).mock.calls[0][1];
  await assistant.deleteConversation('deleted');
  expect(signal?.aborted).toBe(true);
  response.resolve(history('deleted')); await pending;
  expect(assistant.activeConversation.value.id).toBe('original');
});

test('deleting the active chat preserves selection of another conversation', async () => {
  const response = deferred<ConversationHistoryResponse>();
  jest.mocked(fetchConversationHistory).mockReturnValueOnce(response.promise);
  jest.mocked(deleteConversation).mockResolvedValueOnce();
  const pending = assistant.selectConversation('next');
  const signal = jest.mocked(fetchConversationHistory).mock.calls[0][1];
  await assistant.deleteConversation('original');
  expect(signal?.aborted).toBe(false);
  response.resolve(history('next')); await pending;
  expect(assistant.activeConversation.value.id).toBe('next');
});
