import { ref } from 'vue';
import { useConnectionBoardLayout } from '@/composables/integrations/useConnectionBoardLayout';
import type { ConnectionScope } from '@/composables/useConnections';
function fixture() {
  const state = ref({
    board: {
      etag: 'revision',
      entries: [
        { instance_id: 'a', hidden: false },
        { instance_id: 'b', hidden: false },
      ],
    },
    instances: [{ id: 'a' }, { id: 'b' }],
  });
  const saveBoard = jest.fn();
  const commit = jest.fn((_name, value) => {
    state.value.board = value;
  });
  const scope = {
    state,
    busy: ref(false),
    api: { saveBoard },
    commit,
    perform: async (action: () => Promise<unknown>) => {
      try {
        return await action();
      } catch {
        return undefined;
      }
    },
  } as unknown as ConnectionScope;
  return { state, saveBoard, commit, layout: useConnectionBoardLayout(scope) };
}
test('cancel discards only the draft and failed saves retain it for retry', async () => {
  const { state, saveBoard, layout } = fixture();
  layout.begin();
  layout.move('a', 1);
  layout.hide('a');
  expect(state.value.board.entries[0].hidden).toBe(false);
  saveBoard.mockRejectedValue(new Error('conflict'));
  expect(await layout.save()).toBe(false);
  expect(layout.editing.value).toBe(true);
  expect(layout.entries.value[1]).toEqual({ instance_id: 'a', hidden: true });
  layout.cancel();
  expect(layout.visible.value.map((e) => e.instance_id)).toEqual(['a', 'b']);
});
test('save uses the captured revision and restore preserves all other placements', async () => {
  const { saveBoard, layout } = fixture();
  layout.begin();
  layout.hide('a');
  saveBoard.mockImplementation(async (entries) => ({ etag: 'next', entries }));
  expect(await layout.save()).toBe(true);
  expect(layout.editing.value).toBe(false);
  expect(saveBoard).toHaveBeenCalledWith(
    [
      { instance_id: 'a', hidden: true },
      { instance_id: 'b', hidden: false },
    ],
    'revision'
  );
  expect(await layout.restore('a')).toBe(true);
  expect(layout.visible.value.map((e) => e.instance_id)).toEqual(['a', 'b']);
});
