import { computed, nextTick, onBeforeUnmount, onMounted } from 'vue';
import { onBeforeRouteLeave } from 'vue-router';
import { RouteName } from '@/types/router';
import { useConnections } from '@/composables/useConnections';
import { useConnectionsCatalog } from './useConnectionsCatalog';
import { useConnectionBoardLayout } from './useConnectionBoardLayout';
import type { ConnectionInstance } from '@/types/integrations.type';

/** One page owns all local reads/observations. Widgets have no timers or transport state. */
export function useConnectionsBoard() {
  const scope = useConnections();
  const { state, store, api, signal, isActive, perform, commit, observe, busy, error } = scope;
  const catalog = useConnectionsCatalog(scope);
  const layout = useConnectionBoardLayout(scope);
  const byId = computed(() => new Map(state.value.instances.map((item) => [item.id, item])));
  /** Reload accepted catalog/layout/instance state using this page's cancellation signal. */
  const load = () => perform(() => store.dispatch('integrations/loadBoard', signal));
  async function readInstances(requestSignal = signal) {
    await store.dispatch('integrations/readInstances', requestSignal);
  }
  /** Admit manual refresh or continue an existing lookup, then observe local progress. */
  async function refresh(item: ConnectionInstance) {
    const operation = await perform(() =>
      item.operation_id ? api.check(item.operation_id, item.etag) : api.refresh(item.id, item.etag)
    );
    if (operation) observe(operation.id, readInstances);
  }
  let projection: ReturnType<typeof setInterval> | undefined;
  const onFocus = () => {
    if (
      isActive() &&
      !busy.value &&
      !layout.editing.value &&
      !catalog.open.value &&
      !document.hidden &&
      state.value.catalog?.available
    )
      void perform(() => readInstances());
  };
  // The shell scrolls its app-main ancestor; do not introduce a competing global scroll policy.
  const scroller = () => document.querySelector<HTMLElement>('.nh-app-main');
  onBeforeRouteLeave((to) => {
    if (to.name === RouteName.Integration) scroller()?.scrollTo({ top: 0 });
  });
  /** Remember ordinary card navigation in session memory; modified/new-tab clicks stay independent. */
  function remember(id: string, event: MouseEvent) {
    if (event.button || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;
    commit('boardReturn', { id, scrollTop: scroller()?.scrollTop ?? 0 });
  }
  async function restoreReturn() {
    const saved = state.value.boardReturn;
    if (!saved || !isActive()) return;
    await nextTick();
    const links = Array.from(document.querySelectorAll<HTMLAnchorElement>('.connection-widget-link'));
    const link = links.find((element) => element.dataset.instanceId === saved.id);
    const fallback = document.querySelector<HTMLElement>('[data-board-add]');
    (link ?? fallback)?.focus({ preventScroll: true });
    scroller()?.scrollTo({ top: saved.scrollTop });
    commit('boardReturn', null);
  }
  /** Explicitly reload after a layout conflict; a successful read abandons the stale local draft. */
  async function reloadLayout() {
    await load();
    if (!error.value) layout.cancel();
  }
  onMounted(async () => {
    await load();
    if (!isActive()) return;
    await restoreReturn();
    // Observe at most one admitted local job. Explicit Check handles provider polling later.
    const pending = state.value.instances.find((item) => item.operation_id);
    if (pending?.operation_id) observe(pending.operation_id, readInstances);
    projection = setInterval(onFocus, 60_000);
    window.addEventListener('focus', onFocus);
  });
  const stopProjection = () => {
    if (projection) clearInterval(projection);
    window.removeEventListener('focus', onFocus);
  };
  signal.addEventListener('abort', stopProjection, { once: true });
  onBeforeUnmount(stopProjection);
  return { state, byId, busy, error, load, refresh, remember, catalog, layout, reloadLayout };
}
