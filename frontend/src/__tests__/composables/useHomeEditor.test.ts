import { computed, ref } from 'vue';
import { useHomeEditor } from '@/composables/useHomeEditor';
import { confirm } from '@/composables/useConfirm';
import type { HomeLayout } from '@/types/home.type';

jest.mock('@/composables/useConfirm', () => ({ confirm: jest.fn() }));
const confirmation = confirm as jest.MockedFunction<typeof confirm>;

function fixture() {
  const accepted = ref<HomeLayout>({
    version: 1,
    revision: 1,
    etag: 'one',
    blocks: [
      {
        id: 'area',
        kind: 'group',
        name: 'Kitchen',
        cards: [{ id: 'light', kind: 'device', device_id: 'lamp', expose: 'state', panel: true }],
      },
      { id: 'garden', kind: 'group', name: 'Garden', cards: [] },
    ],
  });
  const save = jest.fn<Promise<boolean>, [HomeLayout]>().mockResolvedValue(false);
  const editing = ref(false);
  const editor = useHomeEditor({ layout: computed(() => accepted.value), save }, editing);
  return { accepted, save, editing, editor };
}

beforeEach(() => confirmation.mockReset());

test('failed area saves preserve accepted state and dialog; accepted saves close it', async () => {
  const { accepted, save, editor, editing } = fixture();
  editor.begin();
  expect(editing.value).toBe(true);
  editor.renameArea(accepted.value.blocks[0]);
  expect(editor.validateAreaName('Garden')).toBe('An area already has this name.');
  expect(editor.validateAreaName('Kitchen')).toBeNull();
  await editor.saveAreaName('Lounge');
  expect(accepted.value.blocks[0].name).toBe('Kitchen');
  expect(editor.areaDialogOpen.value).toBe(true);
  expect(save.mock.calls[0][0].blocks[0]).toMatchObject({ id: 'area', name: 'Lounge' });
  save.mockResolvedValue(true);
  await editor.saveAreaName('Lounge');
  expect(editor.areaDialogOpen.value).toBe(false);
  editor.finish();
  expect(editing.value).toBe(false);
});

test('card transfer creates one complete proposal while accepted membership remains intact', async () => {
  const { accepted, save, editor } = fixture();
  await editor.transferCard({ id: 'light', index: 0 }, 'garden');
  expect(save).toHaveBeenCalledTimes(1);
  const proposal = save.mock.calls[0][0];
  expect(proposal.blocks[0].cards).toHaveLength(0);
  expect(proposal.blocks[1].cards[0].id).toBe('light');
  expect(accepted.value.blocks[0].cards[0].id).toBe('light');
  expect(accepted.value.blocks[1].cards).toHaveLength(0);
});

test('cancelled deletion performs no save; confirmed deletion removes placements only', async () => {
  const { accepted, save, editor } = fixture();
  confirmation.mockResolvedValueOnce(false).mockResolvedValueOnce(true);
  await editor.removeArea(accepted.value.blocks[0]);
  expect(save).not.toHaveBeenCalled();
  await editor.removeArea(accepted.value.blocks[0]);
  expect(save.mock.calls[0][0].blocks.map((block) => block.id)).toEqual(['garden']);
  expect(accepted.value.blocks).toHaveLength(2);
});

test('standalone deletion requires confirmation and keeps source data untouched', async () => {
  const { accepted, save, editor } = fixture();
  const card = accepted.value.blocks[0].cards[0];
  accepted.value.blocks.push({ id: 'standalone', kind: 'card', name: '', cards: [{ ...card, id: 'placed-light' }] });
  const standalone = accepted.value.blocks[2].cards[0];
  confirmation.mockResolvedValueOnce(false).mockResolvedValueOnce(true);
  await editor.removeStandaloneCard(standalone, 'Light');
  expect(save).not.toHaveBeenCalled();
  await editor.removeStandaloneCard(standalone, 'Light');
  expect(save.mock.calls[0][0].blocks.map((block) => block.id)).toEqual(['area', 'garden']);
  expect(accepted.value.blocks[0].cards[0]).toEqual(card);
});
