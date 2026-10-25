import { connectionsApi, ConnectionError } from '@/services/integrations.service';
afterEach(() => jest.restoreAllMocks());
test('definition save sends a revision and cookie mutation header', async () => {
  const fetch = jest.spyOn(global, 'fetch').mockResolvedValue(
    new Response(
      JSON.stringify({
        data: {
          key: 'waste_collection/example',
          yaml: '',
          document: { name: '' },
          content_hash: 'h',
          active_hash: 'h',
          revision: 1,
          etag: '"h"',
          status: 'disk_error',
          update_available: false,
        },
        error: null,
      }),
      { status: 200 }
    )
  );
  await connectionsApi.saveDefinition('waste_collection/example', 'id: example', '"h"');
  const options = fetch.mock.calls[0][1]!;
  expect(options.credentials).toBe('include');
  expect(options.headers).toEqual(expect.objectContaining({ 'If-Match': '"h"', 'X-NodeHerder-Request': '1' }));
});
test('preserves actionable conflict errors and rejects malformed success', async () => {
  jest
    .spyOn(global, 'fetch')
    .mockResolvedValueOnce(
      new Response(
        JSON.stringify({ data: null, error: { code: 'revision_conflict', message: 'Reload before saving.' } }),
        { status: 409 }
      )
    )
    .mockResolvedValueOnce(new Response(JSON.stringify({ data: {}, error: null }), { status: 200 }));
  await expect(connectionsApi.remove('id', 'etag')).rejects.toMatchObject({ failure: { code: 'revision_conflict' } });
  await expect(connectionsApi.instance('id')).rejects.toThrow();
});
test('401 is a session failure without retaining an upstream body', async () => {
  jest.spyOn(global, 'fetch').mockResolvedValue(new Response('private upstream details', { status: 401 }));
  await expect(connectionsApi.instances()).rejects.toBeInstanceOf(ConnectionError);
});
