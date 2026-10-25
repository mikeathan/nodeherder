import { computed, onBeforeUnmount, ref } from 'vue';
import { useRouter } from 'vue-router';
import { store } from '@/store';
import { createConnectionsApi, ConnectionError } from '@/services/integrations.service';
import { createOperationObserver } from '@/domain/connection-operation-observer';
import type { ConnectionsState } from '@/store/modules/integrations';
import type { ConnectionFailure } from '@/types/integrations.type';
import { RouteName } from '@/types/router';

/** Page/session ownership for authenticated requests, feedback and local progress observation. */
export function useConnections() {
  const router = useRouter();
  const state = computed(() => (store.state as { integrations: ConnectionsState }).integrations);
  const busy = ref(false);
  const error = ref<ConnectionFailure | null>(null);
  const controller = new AbortController();
  const api = createConnectionsApi(controller.signal);
  const generation = state.value.generation;
  const isActive = () => !controller.signal.aborted && state.value.generation === generation;
  let observer: ReturnType<typeof createOperationObserver> | undefined;

  function stop() {
    controller.abort();
    observer?.stop();
  }
  const unwatch = store.watch(() => state.value.generation, stop);
  onBeforeUnmount(() => {
    stop();
    unwatch();
  });

  async function reportFailure(cause: unknown) {
    // Replies from an unmounted page or a previous login cannot affect the new session.
    if (!isActive() || (cause instanceof Error && cause.name === 'AbortError')) return;
    if (cause instanceof ConnectionError && cause.status === 401) {
      store.commit('auth/logout');
      await router.replace({ name: RouteName.Login });
      return;
    }
    error.value =
      cause instanceof ConnectionError
        ? cause.failure
        : {
            code: 'invalid_response',
            message: 'This action could not be completed.',
          };
  }

  function commit(name: string, value: unknown) {
    if (isActive()) store.commit(`integrations/${name}`, value);
  }

  async function perform<T>(action: () => Promise<T>): Promise<T | undefined> {
    if (!isActive() || busy.value) return;
    busy.value = true;
    error.value = null;
    try {
      const result = await action();
      return isActive() ? result : undefined;
    } catch (cause) {
      await reportFailure(cause);
      return undefined;
    } finally {
      busy.value = false;
    }
  }

  function observe(id: string, refresh: (signal: AbortSignal) => Promise<unknown>) {
    if (!isActive()) return;
    observer?.stop();
    observer = createOperationObserver({
      read: (operationId, signal) => api.operation(operationId, signal),
      refresh,
      onError: (cause) => void reportFailure(cause),
    });
    observer.start(id);
  }

  return { store, router, state, api, busy, error, perform, commit, observe, isActive, signal: controller.signal };
}
export type ConnectionScope = ReturnType<typeof useConnections>;
