import { store } from '@/store';
import { useActivityFeed } from '@/composables/useActivityFeed';

test('logout clears sensitive modules and activity while keeping the feed preference', () => {
  const activity = useActivityFeed(); activity.setEnabled(true);
  store.commit('hub/setDevices', [{ id: 'd', exposes: { temperature: { data: 1 } } }]);
  store.commit('hub/updateDevice', { id: 'd', data: { temperature: 2 } });
  store.commit('assistant/addMessage', { id: 'm', role: 'user', content: 'private', timestamp: 0 });
  store.commit('assistant/setConversations', [{ id: 'c', title: 'private', created_at: '', message_count: 1 }]);
  store.commit('console/add', { level: 'info', message: 'private', timestamp: 0 });
  const old = store.getters['assistant/activeConversation'].id;
  expect(activity.entries.value).toHaveLength(1);
  store.commit('auth/logout');
  expect(store.getters['assistant/activeConversation'].messages).toEqual([]);
  expect(store.getters['assistant/activeConversation'].id).not.toBe(old);
  expect(store.getters['assistant/conversations']).toEqual([]);
  expect(store.getters['console/messages']()).toEqual([]);
  expect(activity.entries.value).toEqual([]);
  expect(activity.enabled.value).toBe(true);
});
