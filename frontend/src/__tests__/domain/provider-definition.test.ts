import { renameProviderRequest, renameProviderCapture } from '@/domain/provider-definition';
import type { ProviderDefinition } from '@/types/integrations.type';

function definition(): ProviderDefinition {
  return {
    schema_version: 'nodeherder.integration/v1',
    integration_type: 'waste_collection',
    id: 'example',
    name: 'Example',
    capability: 'waste_schedule/v1',
    base_url: 'https://example.com',
    auth: { mode: 'none' },
    policy: {
      poll_min_seconds: 5,
      renewal_min_seconds: 86400,
      idempotency_window_seconds: 86400,
      refresh_default_seconds: 0,
      refresh_min_seconds: 60,
    },
    errors: { code: '/problem', rules: [] },
    inputs: {},
    mapping: { metadata: {}, collections: { items: '/collections', fields: { name: '/name', dates: '/dates' } } },
    requests: {
      submit: {
        method: 'POST',
        path: ['lookups'],
        captures: { property: { pointer: '/property_id', type: 'string' } },
      },
      poll: {
        method: 'GET',
        path: ['lookups'],
        captures: { property: { pointer: '/result/property_id', type: 'string' } },
      },
      schedule: {
        method: 'GET',
        path: ['schedules', { ref: 'state.property' }],
        query: { property: { ref: 'state.property' } },
        body: { literal: 'state.property' },
      },
    },
    setup: {
      type: 'search_select',
      search: {
        input: 'postcode',
        request: 'discover',
        options: { items: '/candidates', value: '/id', label: '/label' },
      },
      select: { input: 'property', request: 'submit' },
      status: { request: 'poll', pointer: '/status', pending: ['pending'], done: ['done'], failed: ['failed'] },
    },
    refresh: { request: 'schedule', renewal_request: 'submit' },
  };
}

test('request rename updates selection and renewal without changing the original', () => {
  const original = definition();
  const renamed = renameProviderRequest(original, 'submit', 'retrieve');
  expect(renamed.setup.select.request).toBe('retrieve');
  expect(renamed.refresh.renewal_request).toBe('retrieve');
  expect(renamed.requests.retrieve).toBeDefined();
  expect(original.setup.select.request).toBe('submit');
  expect(renameProviderRequest(original, 'submit', 'poll')).toBe(original);
});

test('capture rename updates every producer and typed use while preserving fixed text', () => {
  const original = definition();
  const renamed = renameProviderCapture(original, 'property', 'reference');
  expect(renamed.requests.submit.captures?.reference).toBeDefined();
  expect(renamed.requests.poll.captures?.reference).toBeDefined();
  expect(renamed.requests.schedule.path[1]).toEqual({ ref: 'state.reference' });
  expect(renamed.requests.schedule.query?.property).toEqual({ ref: 'state.reference' });
  expect(renamed.requests.schedule.body?.literal).toBe('state.property');
  expect(original.requests.schedule.path[1]).toEqual({ ref: 'state.property' });
  expect(renameProviderCapture(original, 'property', 'bad name')).toBe(original);
});
