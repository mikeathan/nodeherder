import { integrationSourceTitle, parseIntegrationReference } from '@/integrations/registry';

test('accepts each declared source kind and rejects unowned sub-sources', () => {
  expect(parseIntegrationReference({ kind: 'weather', instance_id: 'i', resource_id: 'r' })).toEqual({
    kind: 'weather',
    instance_id: 'i',
    resource_id: 'r',
  });
  expect(
    parseIntegrationReference({ kind: 'waste_service', instance_id: 'i', resource_id: 'r', service_id: 's' })
  ).toMatchObject({ kind: 'waste_service', service_id: 's' });
  expect(() =>
    parseIntegrationReference({ kind: 'weather', instance_id: 'i', resource_id: 'r', service_id: 's' })
  ).toThrow('Unexpected integration sub-source');
  expect(() => parseIntegrationReference({ kind: 'unknown', instance_id: 'i', resource_id: 'r' })).toThrow(
    'Unsupported integration source'
  );
});

test('resolves source titles and a generic fallback', () => {
  expect(integrationSourceTitle('weather')).toBe('Weather');
  expect(integrationSourceTitle('weather_alerts')).toBe('Weather alerts');
  expect(integrationSourceTitle('removed_kind')).toBe('Integration result');
});
