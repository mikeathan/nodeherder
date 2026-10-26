import { connectionsApi, createConnectionsApi, ConnectionError } from '@/services/integrations.service';
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

test('page lifetime signals are attached to mutations as well as reads', async () => {
  const controller = new AbortController();
  const fetch = jest
    .spyOn(global, 'fetch')
    .mockResolvedValue(new Response(JSON.stringify({ data: {}, error: null }), { status: 200 }));
  const api = createConnectionsApi(controller.signal);
  await api.remove('id', 'etag');
  expect(fetch.mock.calls[0][1]?.signal).toBe(controller.signal);
  controller.abort();
  expect((fetch.mock.calls[0][1]?.signal as AbortSignal).aborted).toBe(true);
});

test('board save uses authenticated revision transport and runtime validation', async () => {
  const entries = [{ instance_id: 'a', hidden: true }];
  const fetch = jest
    .spyOn(global, 'fetch')
    .mockResolvedValue(new Response(JSON.stringify({ data: { etag: 'next', entries }, error: null }), { status: 200 }));
  const controller = new AbortController();
  expect(await createConnectionsApi(controller.signal).saveBoard(entries, 'previous')).toEqual({
    etag: 'next',
    entries,
  });
  expect(fetch.mock.calls[0][0]).toContain('/integrations/board');
  expect(fetch.mock.calls[0][1]).toMatchObject({
    method: 'PUT',
    credentials: 'include',
    signal: controller.signal,
    body: JSON.stringify({ entries }),
    headers: expect.objectContaining({ 'If-Match': 'previous', 'X-NodeHerder-Request': '1' }),
  });
});
