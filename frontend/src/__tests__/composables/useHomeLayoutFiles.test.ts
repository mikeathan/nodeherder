import { computed, ref } from 'vue';
import { useHomeLayoutFiles } from '@/composables/useHomeLayoutFiles';
import { readHomeExport } from '@/services/home-export.service';
import { confirm } from '@/composables/useConfirm';
import type { HomeLayout } from '@/types/home.type';

const mockUnmounted: (() => void)[] = [];
jest.mock('vue', () => ({ ...jest.requireActual('vue'), onBeforeUnmount: (fn: () => void) => mockUnmounted.push(fn) }));
jest.mock('@/services/home-export.service', () => ({ readHomeExport: jest.fn(), downloadHomeExport: jest.fn() }));
jest.mock('@/composables/useConfirm', () => ({ confirm: jest.fn() }));
const read = readHomeExport as jest.MockedFunction<typeof readHomeExport>;
const confirmation = confirm as jest.MockedFunction<typeof confirm>;
const file = new File(['{}'], 'home.json');

function fixture() {
  const accepted = ref<HomeLayout>({ version: 1, revision: 1, etag: 'one', blocks: [] });
  const active = ref(true);
  const save = jest.fn().mockResolvedValue(true);
  const files = useHomeLayoutFiles({ layout: computed(() => accepted.value), save, isActive: () => active.value });
  return { accepted, active, save, files };
}
beforeEach(() => {
  read.mockReset();
  confirmation.mockReset();
});
afterEach(() => mockUnmounted.splice(0).forEach((stop) => stop()));

test('abandoned file reads cannot open a dialog or save after navigation', async () => {
  const { files, save } = fixture();
  let finish!: (layout: HomeLayout) => void;
  read.mockImplementation(
    () =>
      new Promise((resolve) => {
        finish = resolve;
      })
  );
  const pending = files.importFile(file);
  mockUnmounted.splice(0).forEach((stop) => stop());
  finish({ version: 1, revision: 1, etag: 'one', blocks: [] });
  await pending;
  expect(confirmation).not.toHaveBeenCalled();
  expect(save).not.toHaveBeenCalled();
});

test('an import is never retagged onto a newer accepted layout while confirmation is open', async () => {
  const { accepted, files, save } = fixture();
  read.mockImplementation(async (_file, current) => current);
  let answer!: (accepted: boolean) => void;
  confirmation.mockImplementation(
    () =>
      new Promise((resolve) => {
        answer = resolve;
      })
  );
  const pending = files.importFile(file);
  await Promise.resolve();
  await Promise.resolve();
  accepted.value = { ...accepted.value, revision: 2, etag: 'two' };
  answer(true);
  await pending;
  expect(save).toHaveBeenCalledWith(expect.objectContaining({ etag: 'one', revision: 1 }));
});

test('session loss after file validation suppresses confirmation and errors', async () => {
  const { active, files, save } = fixture();
  read.mockImplementation(async () => {
    active.value = false;
    throw new Error('late failure');
  });
  await files.importFile(file);
  expect(files.error.value).toBeNull();
  expect(confirmation).not.toHaveBeenCalled();
  expect(save).not.toHaveBeenCalled();
});

test('overlapping file selections are serialized and the pending state is released on failure', async () => {
  const { files } = fixture();
  let fail!: (error: Error) => void;
  read.mockImplementation(
    () =>
      new Promise((_resolve, reject) => {
        fail = reject;
      })
  );
  const first = files.importFile(file);
  expect(files.importing.value).toBe(true);
  await files.importFile(file);
  expect(read).toHaveBeenCalledTimes(1);
  fail(new Error('Invalid file'));
  await first;
  expect(files.importing.value).toBe(false);
  expect(files.error.value).toBe('Invalid file');
});
