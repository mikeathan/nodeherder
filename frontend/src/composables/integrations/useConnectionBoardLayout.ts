import { computed, ref } from 'vue';
import type { ConnectionScope } from '@/composables/useConnections';
import type { ConnectionBoardEntry } from '@/types/integrations.type';
import { moveBoardEntry, reconcileBoard } from '@/domain/connection-board';

/** Owns only a layout draft. Failed/conflicting saves never erase unsaved placements. */
export function useConnectionBoardLayout(scope: ConnectionScope) {
  const editing = ref(false);
  const draft = ref<ConnectionBoardEntry[]>([]);
  const expected = ref('');
  const accepted = computed(() =>
    reconcileBoard(
      scope.state.value.board?.entries ?? [],
      scope.state.value.instances.map((item) => item.id)
    )
  );
  const entries = computed(() => (editing.value ? draft.value : accepted.value));
  const visible = computed(() => entries.value.filter((entry) => !entry.hidden));
  const hidden = computed(() => entries.value.filter((entry) => entry.hidden));
  /** Copy accepted placements and capture their revision; nothing is persisted yet. */
  function begin() {
    if (!scope.state.value.board || scope.busy.value) return;
    draft.value = accepted.value.map((entry) => ({ ...entry }));
    expected.value = scope.state.value.board.etag;
    editing.value = true;
  }
  /** Abandon only the local draft and return to accepted shared placements. */
  function cancel() {
    editing.value = false;
    draft.value = [];
  }
  /** Hide a draft widget without deleting its configured address, credentials or resources. */
  function hide(id: string) {
    draft.value = draft.value.map((entry) => (entry.instance_id === id ? { ...entry, hidden: true } : entry));
  }
  /** Move among visible draft entries while retaining hidden membership. */
  function move(id: string, offset: number) {
    draft.value = moveBoardEntry(draft.value, id, offset);
  }
  /** Merge drag-provided visible order into the full draft, retaining hidden entries. */
  function reorder(items: ConnectionBoardEntry[]) {
    let index = 0;
    draft.value = draft.value.map((entry) => (entry.hidden ? entry : { ...items[index++] }));
  }
  /** Atomically save the captured-revision draft; conflicts/failures keep it available for retry. */
  async function save() {
    const result = await scope.perform(() => scope.api.saveBoard(draft.value, expected.value));
    if (result) {
      scope.commit('board', result);
      cancel();
    }
    return !!result;
  }
  /** Persist visibility for an existing configured connection, preserving other placements. */
  async function restore(id: string) {
    const board = scope.state.value.board;
    if (!board) return false;
    const proposal = accepted.value.map((entry) => (entry.instance_id === id ? { ...entry, hidden: false } : entry));
    const result = await scope.perform(() => scope.api.saveBoard(proposal, board.etag));
    if (result) scope.commit('board', result);
    return !!result;
  }
  return { editing, entries, visible, hidden, begin, cancel, hide, move, reorder, save, restore };
}
