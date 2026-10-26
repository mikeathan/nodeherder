import { helpTopics } from '@/domain/integrations-help';
test('help is searchable by user task and explains mapping once per list', () => {
  expect(helpTopics('changed URL').map((t) => t.id)).toContain('repair');
  expect(helpTopics('per bin')[0].id).toBe('fields');
  expect(helpTopics('nonsense string')).toHaveLength(0);
});
