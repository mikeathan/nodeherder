import { COMMON_HELP, matchesHelpTopic } from '@/domain/integrations/help';
import { integrationHelp } from '@/integrations/registry';

const common = (search: string) => COMMON_HELP.filter((topic) => matchesHelpTopic(topic, search));

test('shared help is searchable and does not name a specific integration', () => {
  expect(common('changed URL').map((topic) => topic.id)).toContain('repair');
  expect(common('JSON Pointer')[0].id).toBe('fields');
  expect(common('nonsense string')).toHaveLength(0);
  expect(COMMON_HELP.map((topic) => topic.paragraphs.join(' ')).join(' ')).not.toMatch(
    /bin collection|waste_collection|fuel prices|severe weather/i
  );
});

test('capability help follows the selected provider', () => {
  expect(integrationHelp('alerts', 'nws')[0].paragraphs.join(' ')).toContain('United States');
  expect(integrationHelp('alerts', 'met_office')[0].paragraphs.join(' ')).toContain('United Kingdom');
  expect(integrationHelp('fuel_prices', 'fuel_finder')[0].id).toBe('fuel-prices-setup');
});
