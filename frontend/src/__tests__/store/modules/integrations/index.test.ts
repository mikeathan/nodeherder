import { createStore } from 'vuex';
import { IntegrationsModule, type ConnectionsState } from '@/store/modules/integrations';
import { connectionsApi, ConnectionError } from '@/services/integrations.service';
import type { ConnectionCatalog, ConnectionInstance, ConnectionDefinition } from '@/types/integrations.type';
const catalog: ConnectionCatalog = { available: true, availability_error: null, types: [], providers: [] };
afterEach(() => jest.restoreAllMocks());
function fixture() {
  const logout = jest.fn();
  const store = createStore({
    modules: { integrations: IntegrationsModule, auth: { namespaced: true, state: () => ({}), mutations: { logout } } },
  });
  return { store, logout };
}
test('a late response cannot repopulate private state after logout cleanup', async () => {
  const { store } = fixture();
  let resolve!: (v: ConnectionCatalog) => void;
  jest.spyOn(connectionsApi, 'catalog').mockImplementation(
    () =>
      new Promise((done) => {
        resolve = done;
      })
  );
  jest.spyOn(connectionsApi, 'instances').mockResolvedValue([]);
  const pending = store.dispatch('integrations/load');
  store.commit('integrations/clear');
  resolve(catalog);
  await pending;
  expect((store.state as { integrations: ConnectionsState }).integrations.catalog).toBeNull();
  expect((store.state as { integrations: ConnectionsState }).integrations.drafts).toEqual({});
  expect((store.state as { integrations: ConnectionsState }).integrations.definitions).toEqual({});
});
test('a late 401 from an earlier feature generation cannot sign out a new session', async () => {
  const { store, logout } = fixture();
  let reject!: (e: unknown) => void;
  jest.spyOn(connectionsApi, 'catalog').mockImplementation(
    () =>
      new Promise((_resolve, fail) => {
        reject = fail;
      })
  );
  jest.spyOn(connectionsApi, 'instances').mockResolvedValue([]);
  const pending = store.dispatch('integrations/load');
  store.commit('integrations/clear');
  reject(new ConnectionError({ code: 'unauthorized', message: 'Expired' }, 401));
  await expect(pending).rejects.toBeInstanceOf(ConnectionError);
  expect(logout).not.toHaveBeenCalled();
});

test('an aborted catalog read cannot replace data even when the transport resolves', async () => {
  const { store } = fixture();
  const controller = new AbortController();
  let resolve!: (value: ConnectionCatalog) => void;
  jest.spyOn(connectionsApi, 'catalog').mockImplementation(
    () =>
      new Promise((done) => {
        resolve = done;
      })
  );
  jest.spyOn(connectionsApi, 'instances').mockResolvedValue([]);
  const pending = store.dispatch('integrations/load', controller.signal);
  controller.abort();
  resolve(catalog);
  await pending;
  expect((store.state as { integrations: ConnectionsState }).integrations.catalog).toBeNull();
});

test('draft retention is bounded and revisiting a draft makes it the most recent', () => {
  const { store } = fixture();
  const save = (key: string) =>
    store.commit('integrations/draft', {
      key,
      value: {
        yaml: 'name: example',
        etag: 'etag',
        mode: 'yaml',
        dirty: true,
        scroll: 0,
      },
    });
  for (let index = 0; index < 32; index++) save(`provider_${index}`);
  save('provider_0');
  save('provider_32');
  const drafts = (store.state as { integrations: ConnectionsState }).integrations.drafts;
  expect(Object.keys(drafts)).toHaveLength(32);
  expect(drafts.provider_0).toBeDefined();
  expect(drafts.provider_1).toBeUndefined();
  expect(drafts.provider_32).toBeDefined();
});

test('accepted definition caches are bounded, keyed, explicitly removable and session-scoped', () => {
  const { store } = fixture();
  // This store boundary only owns DTO membership; payload validation is tested at API ingress.
  const put = (key: string) => store.commit('integrations/definition', { key } as ConnectionDefinition);
  for (let index = 0; index < 32; index++) put(`provider_${index}`);
  put('provider_0');
  put('provider_32');
  const state = (store.state as { integrations: ConnectionsState }).integrations;
  expect(Object.keys(state.definitions)).toHaveLength(32);
  expect(state.definitions.provider_0.key).toBe('provider_0');
  expect(state.definitions.provider_1).toBeUndefined();
  store.commit('integrations/forgetDefinition', 'provider_0');
  expect(state.definitions.provider_0).toBeUndefined();
  store.commit('integrations/clear');
  expect(state.definitions).toEqual({});
});

test('a late board load cannot repopulate layout or return state after cleanup', async () => {
  const { store } = fixture();
  let resolve!: (value: { etag: string; entries: [] }) => void;
  jest.spyOn(connectionsApi, 'catalog').mockResolvedValue(catalog);
  jest.spyOn(connectionsApi, 'instances').mockResolvedValue([]);
  jest.spyOn(connectionsApi, 'board').mockImplementation(
    () =>
      new Promise((done) => {
        resolve = done;
      })
  );
  const pending = store.dispatch('integrations/loadBoard');
  await Promise.resolve();
  store.commit('integrations/boardReturn', { id: 'private-instance', scrollTop: 50 });
  store.commit('integrations/clear');
  resolve({ etag: 'revision', entries: [] });
  await pending;
  const state = (store.state as { integrations: ConnectionsState }).integrations;
  expect(state.board).toBeNull();
  expect(state.boardReturn).toBeNull();
});

test('a superseded local projection read cannot replace the newer result', async () => {
  const { store } = fixture();
  let finishFirst!: (items: ConnectionInstance[]) => void;
  let finishSecond!: (items: ConnectionInstance[]) => void;
  jest
    .spyOn(connectionsApi, 'instances')
    .mockImplementationOnce(
      () =>
        new Promise((done) => {
          finishFirst = done;
        })
    )
    .mockImplementationOnce(
      () =>
        new Promise((done) => {
          finishSecond = done;
        })
    );
  const first = store.dispatch('integrations/readInstances');
  const second = store.dispatch('integrations/readInstances');
  finishSecond([{ id: 'newer' } as ConnectionInstance]);
  await second;
  finishFirst([{ id: 'older' } as ConnectionInstance]);
  await first;
  expect((store.state as { integrations: ConnectionsState }).integrations.instances[0].id).toBe('newer');
});
