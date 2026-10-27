import { createStore } from 'vuex';
import { HomeModule, type HomeState } from '@/store/modules/home';
import { homeApi, HomeError } from '@/services/home.service';
import type { HomeLayout } from '@/types/home.type';
import type { DashboardGroups } from '@/types/settings.type';

const layout = (revision: number): HomeLayout => ({ version: 1, revision, etag: `"${revision}"`, blocks: [] });
const fixture = () =>
  createStore({
    modules: {
      home: HomeModule,
      hub: {
        namespaced: true,
        state: () => ({ groups: {} as DashboardGroups }),
        mutations: {
          setDashboardGroups(state, groups: DashboardGroups) {
            state.groups = groups;
          },
        },
      },
    },
  });
const state = (store: ReturnType<typeof fixture>) => (store.state as { home: HomeState }).home;
afterEach(() => jest.restoreAllMocks());

test('a late read cannot replace a newer accepted save', async () => {
  const store = fixture();
  const signal = new AbortController().signal;
  let finish!: (value: HomeLayout) => void;
  jest.spyOn(homeApi, 'read').mockImplementation(
    () =>
      new Promise((resolve) => {
        finish = resolve;
      })
  );
  jest.spyOn(homeApi, 'save').mockResolvedValue(layout(2));
  const pending = store.dispatch('home/read', signal);
  await store.dispatch('home/save', { proposal: layout(1), signal });
  finish(layout(1));
  await pending;
  expect(state(store).layout).toEqual(layout(2));
});

test.each(['logout', 'abort'])('late replies after %s cannot publish private state', async (mode) => {
  const store = fixture();
  const controller = new AbortController();
  let finish!: (value: HomeLayout) => void;
  jest.spyOn(homeApi, 'save').mockImplementation(
    () =>
      new Promise((resolve) => {
        finish = resolve;
      })
  );
  const pending = store.dispatch('home/save', { proposal: layout(1), signal: controller.signal });
  if (mode === 'logout') store.commit('home/clear');
  else controller.abort();
  finish(layout(2));
  await pending;
  expect(state(store).layout).toBeNull();
});

test('a revision conflict preserves the accepted layout and rejects without retry', async () => {
  const store = fixture();
  store.commit('home/accepted', layout(1));
  const save = jest
    .spyOn(homeApi, 'save')
    .mockRejectedValue(new HomeError({ code: 'conflict', message: 'Reload layout.' }, 409));
  await expect(
    store.dispatch('home/save', { proposal: layout(1), signal: new AbortController().signal })
  ).rejects.toBeInstanceOf(HomeError);
  expect(state(store).layout).toEqual(layout(1));
  expect(save).toHaveBeenCalledTimes(1);
});

test('accepted layout refreshes the legacy view without private integration references', async () => {
  const store = fixture();
  const accepted: HomeLayout = {
    ...layout(2),
    blocks: [
      {
        id: 'area',
        kind: 'group',
        name: 'Kitchen',
        cards: [
          { id: 'lamp', kind: 'device', device_id: 'light', expose: 'state', panel: true },
          {
            id: 'private',
            kind: 'waste_summary',
            instance_id: 'private-instance',
            resource_id: 'calendar',
            panel: true,
          },
        ],
      },
    ],
  };
  jest.spyOn(homeApi, 'read').mockResolvedValue(accepted);
  await store.dispatch('home/read', new AbortController().signal);
  const groups = (store.state as { hub: { groups: DashboardGroups } }).hub.groups;
  expect(groups).toEqual({
    Kitchen: { name: 'Kitchen', order: 0, deviceGroup: { light: { deviceId: 'light', exposes: ['state'] } } },
  });
  expect(JSON.stringify(groups)).not.toContain('private-instance');
});
