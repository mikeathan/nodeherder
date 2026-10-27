import { computed, onBeforeUnmount, onMounted, readonly, ref, type Ref } from 'vue';
import { useRouter } from 'vue-router';
import { store } from '@/store';
import type { HomeState } from '@/store/modules/home';
import type { ConnectionsState } from '@/store/modules/integrations';
import type { HomeLayout } from '@/types/home.type';
import type { DashboardGroups } from '@/types/settings.type';
import { HomeError } from '@/services/home.service';
import { ConnectionError } from '@/services/integrations.service';
import { legacyHome } from '@/domain/home-compatibility';
import { homeIntegrationSources, type HomeSourcePresentation } from '@/domain/home-cards';
import { RouteName } from '@/types/router';

interface HomeLayoutOptions {
  pauseRefresh?: Readonly<Ref<boolean>>;
}

const PROJECTION_REFRESH_MS = 60_000;

/** Own one Home/panel request lifetime and accepted-state access. Editor state is supplied only
 * to pause automatic refresh; cards and dialogs never own transport, source truth or polling. */
export function useHomeLayout(options: HomeLayoutOptions = {}) {
  const router = useRouter();
  const state = computed(() => (store.state as { home: HomeState }).home);
  const integrations = computed(() => (store.state as { integrations: ConnectionsState }).integrations);
  const controller = new AbortController();
  const generation = state.value.generation;
  const busy = ref(false);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const sourceError = ref<string | null>(null);
  let projectionTimer: ReturnType<typeof setInterval> | undefined;
  let readingSources = false;

  const fallback = computed(() => legacyHome(store.getters['hub/dashboardGroups']() as DashboardGroups));
  const layout = computed(() => state.value.layout ?? fallback.value);
  const writable = computed(() => Boolean(state.value.layout?.etag) && isActive());
  const sources = computed<ReadonlyMap<string, HomeSourcePresentation>>(() =>
    homeIntegrationSources(integrations.value.instances)
  );

  /** Whether this mounted page still belongs to the login generation that created it. */
  function isActive(): boolean {
    return !controller.signal.aborted && generation === state.value.generation;
  }

  async function reportFailure(cause: unknown, destination: Ref<string | null>) {
    if (!isActive()) return;
    if (cause instanceof Error && cause.name === 'AbortError') return;
    const unauthorized = (cause instanceof HomeError || cause instanceof ConnectionError) && cause.status === 401;
    if (unauthorized) {
      store.commit('auth/logout');
      await router.replace({ name: RouteName.Login });
      return;
    }
    destination.value = cause instanceof Error ? cause.message : 'Home could not be loaded or saved.';
  }

  /** Read local layout and projections independently. A source failure keeps device Home usable. */
  async function load() {
    if (!isActive() || busy.value || loading.value || readingSources) return;
    loading.value = true;
    error.value = null;
    sourceError.value = null;
    const [layoutResult, sourceResult] = await Promise.allSettled([
      store.dispatch('home/read', controller.signal),
      store.dispatch('integrations/readInstances', controller.signal),
    ]);
    if (!isActive()) return;
    if (layoutResult.status === 'rejected') await reportFailure(layoutResult.reason, error);
    if (sourceResult.status === 'rejected') await reportFailure(sourceResult.reason, sourceError);
    loading.value = false;
  }

  /** Save one revision-fenced proposal. Failure retains accepted state and returns false; no
   * implicit retries or source/device commands occur. The caller owns dialog/drag reconciliation. */
  async function save(proposal: HomeLayout): Promise<boolean> {
    if (!isActive() || !writable.value || busy.value || loading.value) return false;
    busy.value = true;
    error.value = null;
    try {
      await store.dispatch('home/save', { proposal, signal: controller.signal });
      return isActive();
    } catch (cause) {
      await reportFailure(cause, error);
      return false;
    } finally {
      if (isActive()) busy.value = false;
    }
  }

  function mayRefresh(): boolean {
    return isActive() && !document.hidden && !options.pauseRefresh?.value;
  }

  function reloadOnFocus() {
    if (mayRefresh()) void load();
  }

  async function refreshProjections() {
    if (!mayRefresh() || busy.value || loading.value || readingSources) return;
    readingSources = true;
    try {
      // This is a local projection read, never a provider refresh. Vuex fences late replies.
      await store.dispatch('integrations/readInstances', controller.signal);
      if (isActive()) sourceError.value = null;
    } catch (cause) {
      await reportFailure(cause, sourceError);
    } finally {
      readingSources = false;
    }
  }

  function stop() {
    controller.abort();
    if (projectionTimer !== undefined) clearInterval(projectionTimer);
    projectionTimer = undefined;
    window.removeEventListener('focus', reloadOnFocus);
    document.removeEventListener('visibilitychange', reloadOnFocus);
  }

  const unwatchSession = store.watch(() => state.value.generation, stop);
  onMounted(() => {
    if (!isActive()) return;
    void load();
    projectionTimer = setInterval(refreshProjections, PROJECTION_REFRESH_MS);
    window.addEventListener('focus', reloadOnFocus);
    document.addEventListener('visibilitychange', reloadOnFocus);
  });
  onBeforeUnmount(() => {
    stop();
    unwatchSession();
  });

  return {
    layout,
    sources,
    writable,
    busy: readonly(busy),
    loading: readonly(loading),
    error: readonly(error),
    sourceError: readonly(sourceError),
    load,
    save,
    isActive,
  };
}

export type HomeLayoutScope = ReturnType<typeof useHomeLayout>;
