import { computed, onScopeDispose, ref, watch, type WatchSource } from 'vue';
import { store } from '@/store';
import { acquireHistory } from '@/services/history.service';
import { historyKey, type HistoryEntry, type HistoryRequest } from '@/domain/history/types';

/** Owns the view's subscription, never the confirmed device state. */
export function useHistoryQuery(source: WatchSource<HistoryRequest | null>) {
  const key = ref('');
  const revision = ref(0);
  let release: (() => void) | null = null;
  let generation = 0;
  let debounce: ReturnType<typeof setTimeout> | null = null;
  const entry = computed<HistoryEntry | undefined>(() => store.getters['metrics/history'](key.value));
  function dispose() {
    generation++;
    release?.();
    release = null;
    if (debounce) clearTimeout(debounce);
  }
  function isCurrent(current: number, epoch: number) {
    // View changes and session resets are independent; both invalidate late responses.
    return current === generation && epoch === store.getters['metrics/historyEpoch'];
  }
  async function load(request: HistoryRequest | null) {
    dispose();
    if (!request) {
      key.value = '';
      return;
    }
    key.value = historyKey(request);
    const currentKey = key.value;
    const current = generation;
    const epoch = store.getters['metrics/historyEpoch'] as number;
    store.commit('metrics/historySet', { key: currentKey, entry: { status: 'loading', series: [], error: '' } });
    try {
      const handle = acquireHistory(request);
      release = handle.release;
      const series = await handle.promise;
      if (isCurrent(current, epoch))
        store.commit('metrics/historySet', { key: currentKey, entry: { status: 'ready', series, error: '' } });
    } catch (error) {
      if (isCurrent(current, epoch))
        store.commit('metrics/historySet', {
          key: currentKey,
          entry: { status: 'error', series: [], error: String((error as Error).message) },
        });
    }
  }
  watch(
    [source, revision],
    ([request]) => {
      dispose();
      key.value = request ? historyKey(request) : '';
      if (request)
        store.commit('metrics/historySet', { key: key.value, entry: { status: 'loading', series: [], error: '' } });
      debounce = setTimeout(() => {
        debounce = null;
        void load(request);
      }, 300);
    },
    { immediate: true }
  );
  onScopeDispose(dispose);
  return { entry, refresh: () => revision.value++ };
}
