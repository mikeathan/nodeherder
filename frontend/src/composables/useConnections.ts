import { computed, onBeforeUnmount, ref } from 'vue';
import { store } from '@/store';
import { useRouter } from 'vue-router';
import { connectionsApi, ConnectionError } from '@/services/integrations.service';
import type { ConnectionsState } from '@/store/modules/integrations';
import type { ConnectionFailure } from '@/types/integrations.type';
import { RouteName } from '@/types/router';

// Page-owned reads/timers stop on unmount/logout; shared drafts live only in feature Vuex memory.
export function useConnections() {
  const router = useRouter();
  const state = computed(() => (store.state as { integrations: ConnectionsState }).integrations);
  const busy = ref(false);
  const error = ref<ConnectionFailure | null>(null);
  const controller = new AbortController();
  const generation = state.value.generation;
  let alive = true;
  let timer: ReturnType<typeof setTimeout> | undefined;
  const stop = () => {
    alive = false;
    controller.abort();
    if (timer) clearTimeout(timer);
  };
  const unwatch = store.watch(() => state.value.generation, stop);
  onBeforeUnmount(() => {
    stop();
    unwatch();
  });
  function commit(name: string, value: unknown) {
    if (alive && state.value.generation === generation) store.commit(`integrations/${name}`, value);
  }
  async function perform<T>(fn: () => Promise<T>): Promise<T | undefined> {
    if (!alive || busy.value) return;
    busy.value = true;
    error.value = null;
    try {
      const result = await fn();
      return alive ? result : undefined;
    } catch (e) {
      if (e instanceof Error && e.name === 'AbortError') return;
      if (e instanceof ConnectionError && e.status === 401) {
        if (!alive && store.getters['auth/isAuthenticated']()) return;
        store.commit('auth/logout');
        await router.replace({ name: RouteName.Login });
        return;
      }
      if (alive)
        error.value =
          e instanceof ConnectionError
            ? e.failure
            : {
                code: 'invalid_response',
                message: e instanceof Error ? e.message : 'This action could not be completed.',
              };
      return undefined;
    } finally {
      busy.value = false;
    }
  }
  // Observe local execution only. A pending provider lookup requires an explicit Check status.
  function observe(id: string, refresh: () => Promise<unknown>) {
    if (timer) clearTimeout(timer);
    const deadline = Date.now() + 30_000;
    const tick = async () => {
      if (!alive) return;
      try {
        const operation = await connectionsApi.operation(id, controller.signal);
        await refresh();
        if (alive && ['queued', 'running'].includes(operation.status) && Date.now() < deadline)
          timer = setTimeout(tick, 700);
      } catch (e) {
        if (e instanceof ConnectionError && e.status === 401) {
          if (!store.getters['auth/isAuthenticated']() || alive) {
            store.commit('auth/logout');
            await router.replace({ name: RouteName.Login });
          }
        } else if (alive && !(e instanceof Error && e.name === 'AbortError')) {
          error.value =
            e instanceof ConnectionError
              ? e.failure
              : { code: 'invalid_response', message: 'The operation status could not be read.' };
        }
      }
    };
    void tick();
  }
  return { store, router, state, busy, error, perform, commit, observe, signal: controller.signal };
}
