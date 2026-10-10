<script setup lang="ts">
  /*
   * Large touch target for one device in panel mode. Devices with an on/off state toggle on
   * tap (same command path as Home); others link to the device page and show their value.
   */
  import { computed } from 'vue';
  import { RouteName } from '@/types/router';
  import { binaryOffValue, binaryOnValue, exposeKind, exposeRange, formatExposeValue, isBinaryOn, stateExposeOf } from '@/domain/exposes';
  import { findDevice, statusOf } from '@/composables/useHub';
  import { useDeviceCommand } from '@/composables/useDeviceCommand';
  import { deviceIcon, exposeIcon } from '@/components/ui/icons';
  import UiIcon from '@/components/ui/UiIcon.vue';

  const props = defineProps<{ deviceId: string; expose: string; label: string }>();
  const { send, isPending, canControl } = useDeviceCommand();

  const device = computed(() => findDevice(props.deviceId));
  const entity = computed(() => device.value?.exposes[props.expose]);
  const status = computed(() => (device.value ? statusOf(device.value) : 'missing'));
  const sw = computed(() => {
    const e = entity.value;
    if (!device.value || !e) return undefined;
    return e.type === 'binary' && e.access_mode !== 'read' ? e : stateExposeOf(device.value);
  });
  const on = computed(() => status.value === 'online' && !!sw.value && isBinaryOn(sw.value));
  const pending = computed(() => !!sw.value && isPending(props.deviceId, sw.value.name));
  const brightness = computed(() => device.value?.exposes['brightness']);
  const ring = computed(() => {
    const b = brightness.value;
    if (!b || !on.value || typeof b.data !== 'number') return null;
    return Math.round((b.data / (exposeRange(b)?.max || 254)) * 100);
  });
  const text = computed(() => {
    if (status.value === 'missing') return 'Removed';
    if (status.value === 'offline') return 'Offline';
    if (status.value === 'disabled') return 'Disabled';
    if (pending.value) return 'Sending…';
    if (sw.value) return (on.value ? 'On' : 'Off') + (ring.value !== null && brightness.value ? ` · ${formatExposeValue(brightness.value)}` : '');
    return entity.value ? formatExposeValue(entity.value) : '—';
  });
  const icon = computed(() => (sw.value && device.value ? deviceIcon(device.value) : exposeIcon(props.expose, on.value)));

  function toggle() {
    const target = sw.value;
    if (!target || !canControl(device.value)) return;
    send(device.value, target.name, isBinaryOn(target) ? binaryOffValue(target) : binaryOnValue(target));
  }
</script>

<template>
  <button
    v-if="sw"
    type="button"
    class="pb"
    :class="[`kind-${exposeKind(expose)}`, { 'is-on': on, 'is-pending': pending, 'is-off': status !== 'online' }]"
    :aria-pressed="on"
    :disabled="!canControl(device)"
    :aria-label="`${label}: ${text}. Tap to turn ${on ? 'off' : 'on'}`"
    @click="toggle">
    <span class="pb-ic">
      <svg v-if="ring !== null" class="pb-ring" viewBox="0 0 100 100" aria-hidden="true"><circle cx="50" cy="50" r="44" pathLength="100" :stroke-dasharray="`${ring} 100`" /></svg>
      <UiIcon :path="icon" />
    </span>
    <b>{{ label }}</b>
    <span class="pb-txt">{{ text }}</span>
  </button>
  <RouterLink v-else :to="{ name: RouteName.DevicePage, params: { id: deviceId } }" class="pb" :class="[`kind-${exposeKind(expose)}`, { 'is-off': status !== 'online' }]">
    <span class="pb-ic"><UiIcon :path="icon" /></span>
    <b>{{ label }}</b>
    <span class="pb-txt">{{ text }}</span>
  </RouterLink>
</template>

<style scoped>
  .pb {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 0.25rem;
    min-height: 9.5rem;
    padding: 1.1rem;
    border-radius: var(--nh-radius-l);
    border: var(--nh-border-w) solid var(--nh-border);
    background: var(--nh-surface);
    color: var(--nh-text);
    text-align: left;
    text-decoration: none;
    cursor: pointer;
    font: inherit;
    transition:
      background var(--nh-motion),
      transform 80ms;
    min-width: 0;
  }
  .pb:active:not(:disabled) {
    transform: scale(0.98);
  }
  .pb:disabled {
    cursor: default;
  }
  .pb b {
    font-size: 1.1rem;
    font-weight: 500;
    margin-top: auto;
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .pb-txt {
    color: var(--nh-text-muted);
  }
  .pb-ic {
    position: relative;
    display: grid;
    place-items: center;
    width: 3.6rem;
    height: 3.6rem;
    border-radius: 50%;
    background: var(--nh-surface-2);
    color: var(--nh-kc, var(--nh-text-muted));
  }
  .pb-ic .nh-ic {
    width: 1.7rem;
    height: 1.7rem;
  }
  .pb-ring {
    position: absolute;
    inset: -6px;
    width: calc(100% + 12px);
    height: calc(100% + 12px);
    transform: rotate(-90deg);
  }
  .pb-ring circle {
    fill: none;
    stroke: #2b1a00;
    stroke-width: 6;
    stroke-linecap: round;
    opacity: 0.5;
  }
  .pb.is-on {
    background: var(--dk-lamp);
    border-color: transparent;
    color: #2b1a00;
    box-shadow: 0 10px 40px color-mix(in srgb, var(--dk-lamp) 35%, transparent);
  }
  .pb.is-on .pb-txt {
    color: #2b1a00;
    opacity: 0.8;
  }
  .pb.is-on .pb-ic {
    background: rgb(255 255 255 / 0.45);
    color: #2b1a00;
  }
  .pb.is-pending {
    animation: nh-pulse 0.8s infinite;
  }
  .pb.is-off {
    opacity: 0.6;
  }
  @media (max-width: 760px) {
    .pb {
      min-height: 8rem;
      padding: 0.9rem;
    }
  }
</style>
