/**
 * MCP server process controls for the settings section.
 *
 * The hub acknowledges a start/stop/restart by pushing a new `mcpStatus` frame, so the
 * button must show a pending state until that frame arrives rather than guessing with a
 * fixed delay. Every timer and watcher is owned here and released when the effect scope
 * (the component instance) is disposed.
 */
import { computed, onScopeDispose, ref, watch } from 'vue';
import { store } from '@/store';
import type { MCPStatusType } from '@/types/settings.type';

/** One button at a time may be pending; long enough for a real restart, short enough to fail. */
export const MCP_STATUS_TIMEOUT_MS = 15_000;

export const MCP_ACTIONS = {
  start: 'start',
  stop: 'stop',
  restart: 'restart',
} as const;

export type McpAction = (typeof MCP_ACTIONS)[keyof typeof MCP_ACTIONS];

const MCP_ACTION_DISPATCH: Readonly<Record<McpAction, string>> = {
  [MCP_ACTIONS.start]: 'hub/startMCP',
  [MCP_ACTIONS.stop]: 'hub/stopMCP',
  [MCP_ACTIONS.restart]: 'hub/restartMCP',
};

/** Copy status so a replaced store object is observable with `watch`. */
function snapshotStatus(status: MCPStatusType | null): MCPStatusType | null {
  return status ? { ...status } : null;
}

/**
 * Controls and observable state for the MCP server card.
 *
 * `pending` names the action in flight (or `null`), `error` explains a timeout, and
 * `refresh` re-reads status without changing the server process.
 */
export function useMcpControls() {
  const status = computed<MCPStatusType | null>(() => snapshotStatus(store.getters['hub/mcpStatus']()));
  const enabled = computed(() => store.getters['hub/mcpConfig']() === true);
  const pending = ref<McpAction | null>(null);
  const error = ref('');

  let timeout: ReturnType<typeof setTimeout> | null = null;

  function clearPending() {
    if (timeout !== null) {
      clearTimeout(timeout);
      timeout = null;
    }
    pending.value = null;
  }

  /**
   * A new status is the hub's answer, so it clears both the pending action and any
   * timeout message left by an earlier attempt.
   */
  const stopWatching = watch(status, () => {
    if (pending.value !== null) clearPending();
    if (error.value) error.value = '';
  });

  onScopeDispose(() => {
    stopWatching();
    clearPending();
  });

  /** Ask the hub for its current MCP status; used on mount and by the Refresh action. */
  function refresh() {
    store.dispatch('hub/loadMCPStatus');
  }

  /**
   * Dispatch one process action and wait for the hub's status frame.
   *
   * Ignores a second request while another action is pending so repeated taps cannot
   * queue several process changes.
   */
  async function run(action: McpAction) {
    if (pending.value !== null) return;
    error.value = '';
    pending.value = action;
    timeout = setTimeout(() => {
      timeout = null;
      pending.value = null;
      error.value = 'The hub did not report a new MCP status. Try Refresh to check again.';
    }, MCP_STATUS_TIMEOUT_MS);
    await store.dispatch(MCP_ACTION_DISPATCH[action]);
    refresh();
  }

  return { status, enabled, pending, error, refresh, run };
}
