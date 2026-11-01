import { onBeforeUnmount, ref } from 'vue';
import { ESPHOME_MESSAGES, ESPHOME_POLL_INTERVAL_MS } from '@/contracts/esphome';
import { listESPHomeDevices, type ESPHomeEnrollment } from '@/services/esphome';

/** Owns one dialog's enrollment polling and cancellable read. Mutations stay with
 * the caller; failed refreshes retain the last loaded rows. Scope disposal stops I/O. */
export function useESPHomeEnrollmentList() {
  const items = ref<ESPHomeEnrollment[]>([]);
  const loading = ref(false);
  const loaded = ref(false);
  const loadError = ref('');
  let timer: ReturnType<typeof setInterval> | undefined;
  let request: AbortController | undefined;

  /** Loads redacted settings once, coalescing reads and ignoring retired replies. */
  async function refresh() {
    if (loading.value) return;
    loading.value = true;
    const current = new AbortController();
    request = current;
    try {
      const result = await listESPHomeDevices(current.signal);
      if (request !== current) return;
      items.value = result;
      loaded.value = true;
      loadError.value = '';
    } catch (cause) {
      if (request === current && !(cause instanceof Error && cause.name === 'AbortError')) {
        loadError.value = ESPHOME_MESSAGES.loadFailed;
      }
    } finally {
      if (request === current) { loading.value = false; request = undefined; }
    }
  }

  /** Cancels only the current read, leaving loaded rows and polling ownership intact. */
  function cancelLoad() {
    request?.abort();
    request = undefined;
    loading.value = false;
  }

  /** Stops the polling timer and retires its current read without clearing rows. */
  function stop() {
    clearInterval(timer);
    timer = undefined;
    cancelLoad();
  }

  /** Starts one polling timer; the caller decides when mutations make a read inappropriate. */
  function start(canRefresh: () => boolean) {
    stop();
    void refresh();
    timer = setInterval(() => { if (canRefresh()) void refresh(); }, ESPHOME_POLL_INTERVAL_MS);
  }

  /** Clears presentation data after the dialog closes; call stop before resetting. */
  function reset() {
    items.value = [];
    loaded.value = false;
    loadError.value = '';
  }

  onBeforeUnmount(stop);
  return { items, loading, loaded, loadError, refresh, cancelLoad, start, stop, reset };
}
