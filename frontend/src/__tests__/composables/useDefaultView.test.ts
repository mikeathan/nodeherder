import { RouteName } from '@/types/router';

const getItem = jest.fn();
const setItem = jest.fn();
const removeItem = jest.fn();
jest.mock('@/utils/storage', () => ({ safeStorage: { getItem, setItem, removeItem } }));

async function preference(stored: string | null = null) {
  getItem.mockReturnValue(stored);
  return import('@/composables/useDefaultView');
}

beforeEach(() => {
  jest.resetModules();
  jest.clearAllMocks();
});

test('loads a valid saved view and shares changes between consumers', async () => {
  const { useDefaultView, DEFAULT_VIEW_KEY } = await preference('panel');
  const first = useDefaultView();
  const second = useDefaultView();
  expect(first.preferred.value).toBe(RouteName.Panel);
  first.toggle(RouteName.Overview);
  expect(second.preferred.value).toBe(RouteName.Overview);
  expect(setItem).toHaveBeenCalledWith(DEFAULT_VIEW_KEY, RouteName.Overview);
  second.toggle(RouteName.Overview);
  expect(first.preferred.value).toBeNull();
  expect(removeItem).toHaveBeenCalledWith(DEFAULT_VIEW_KEY);
});

test.each([null, 'unknown', '/panel', 'editor'])('ignores invalid stored view %p', async (stored) => {
  const { useDefaultView } = await preference(stored);
  expect(useDefaultView().preferred.value).toBeNull();
});

test('does not persist unsupported routes', async () => {
  const { useDefaultView } = await preference();
  const view = useDefaultView();
  view.toggle(RouteName.Editor);
  expect(view.preferred.value).toBeNull();
  expect(setItem).not.toHaveBeenCalled();
  expect(removeItem).not.toHaveBeenCalled();
});
