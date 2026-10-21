/*
 * Zigbee permit join (bridgePermitJoin command). The bridge reports its state in
 * bridgeConfig messages; the countdown is derived from the reported duration and the moment
 * this tab saw it start (the hub does not send a start time).
 */
import { computed, onScopeDispose, ref, watch } from 'vue';
import { store } from '@/store';
import { BridgeSettingsType } from '@/types/settings.type';

export const PERMIT_JOIN_SECONDS = 120;

const startedAt = ref<number | null>(null);
const now = ref(Date.now());

const bridge = () => store.getters['hub/bridge']() as BridgeSettingsType | undefined;

/** The application shell owns tracking, including while panel mode hides its consumers. */
export function usePermitJoinTracking() {
  watch(
    () => bridge()?.permitJoin === true,
    (active, _previous, cleanup) => {
      startedAt.value = active ? Date.now() : null;
      now.value = Date.now();
      if (active) {
        const ticker = setInterval(() => (now.value = Date.now()), 1000);
        cleanup(() => clearInterval(ticker));
      }
    },
    { immediate: true }
  );
  onScopeDispose(() => { startedAt.value = null; });
}

export function usePermitJoin() {
  const active = computed(() => bridge()?.permitJoin === true);
  const duration = computed(() => bridge()?.maxTimeAllowed?.value || PERMIT_JOIN_SECONDS);
  const remaining = computed(() => {
    if (!active.value || startedAt.value === null) return 0;
    return Math.max(0, duration.value - Math.floor((now.value - startedAt.value) / 1000));
  });
  const label = computed(() => {
    const s = remaining.value;
    return `${Math.floor(s / 60)}:${String(s % 60).padStart(2, '0')}`;
  });
  return {
    active,
    remaining,
    label,
    start: () => store.dispatch('hub/enablePermitJoin', PERMIT_JOIN_SECONDS),
    stop: () => store.dispatch('hub/disablePermitJoin'),
  };
}
