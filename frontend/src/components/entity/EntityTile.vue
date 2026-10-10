<script setup lang="ts">
  /*
   * Home tile for one device expose (spec 007 US-02). The icon toggles the expose when it is
   * a switch, otherwise the device's on/off state when it has one; the body opens the
   * history view. Missing, offline and disabled devices say so instead of showing values.
   */
  import { computed } from 'vue';
  import { Expose } from '@/types/device';
  import { binaryOffValue, binaryOnValue, exposeKind, exposeLabel, formatExposeValue, isBinaryOn, stateExposeOf } from '@/domain/exposes';
  import { exposeIcon } from '@/components/ui/icons';
  import { findDevice, statusOf } from '@/composables/useHub';
  import { useDeviceCommand } from '@/composables/useDeviceCommand';
  import { emitOpenEntityViewDialog } from '@/contracts/dialog-events';
  import UiIcon from '@/components/ui/UiIcon.vue';

  const props = defineProps<{ deviceId: string; expose: string; editing?: boolean }>();
  const emit = defineEmits<{ (e: 'remove'): void }>();

  const { send, isPending, canControl } = useDeviceCommand();

  const device = computed(() => findDevice(props.deviceId));
  const entity = computed<Expose | undefined>(() => device.value?.exposes[props.expose]);
  const status = computed(() => (device.value ? statusOf(device.value) : 'missing'));

  /** The switch this tile toggles: itself, or the device's state expose. */
  const switchExpose = computed<Expose | undefined>(() => {
    const e = entity.value;
    if (!e || !device.value) return undefined;
    if (e.type === 'binary' && e.access_mode !== 'read') return e;
    return stateExposeOf(device.value);
  });
  const isOn = computed(() => {
    const e = entity.value;
    if (!e || status.value !== 'online') return false;
    if (e.type === 'binary') return isBinaryOn(e);
    return switchExpose.value ? isBinaryOn(switchExpose.value) : false;
  });
  const isAlert = computed(() => ['smoke', 'alarm', 'water_leak', 'tamper', 'device_fault'].includes(props.expose) && isOn.value);
  const pending = computed(() => !!switchExpose.value && isPending(props.deviceId, switchExpose.value.name));
  const toggleable = computed(() => !!switchExpose.value && canControl(device.value) && !props.editing);

  const label = computed(() => entity.value?.description || exposeLabel(props.expose));
  const value = computed(() => {
    if (status.value === 'missing') return 'Device removed';
    if (status.value === 'offline') return 'Offline';
    if (status.value === 'disabled') return 'Disabled';
    if (pending.value) return 'Sending…';
    return entity.value ? formatExposeValue(entity.value) : 'Not reported';
  });

  function toggle() {
    const target = switchExpose.value;
    if (!target || !toggleable.value) return;
    send(device.value, target.name, isBinaryOn(target) ? binaryOffValue(target) : binaryOnValue(target));
  }

  function openHistory() {
    if (props.editing || !device.value) return;
    emitOpenEntityViewDialog(() => {}, { id: props.deviceId, name: props.expose, title: `${exposeLabel(props.expose)} · ${device.value.friendly_name}` });
  }
</script>

<template>
  <div
    class="nh-tile"
    :class="[`kind-${exposeKind(expose)}`, `is-${status}`, { 'is-on': isOn, 'is-alert': isAlert, 'is-pending': pending }]"
    :data-device="deviceId"
    :data-expose="expose">
    <button
      v-if="switchExpose"
      type="button"
      class="nh-tile-ic"
      :aria-label="`${isOn ? 'Turn off' : 'Turn on'} ${device?.friendly_name ?? label}`"
      :aria-pressed="isOn"
      :disabled="!toggleable"
      @click="toggle">
      <UiIcon :path="exposeIcon(expose, isOn)" />
    </button>
    <span v-else class="nh-tile-ic"><UiIcon :path="exposeIcon(expose, isOn)" /></span>
    <button type="button" class="nh-tile-body" :disabled="editing" @click="openHistory">
      <span class="nh-tile-label">{{ label }}</span>
      <span class="nh-tile-value">
        <UiIcon v-if="status === 'offline'" name="offline" />
        <UiIcon v-else-if="status === 'disabled'" name="disabled" />
        {{ value }}
      </span>
      <span class="nh-tile-dev">{{ device?.friendly_name ?? deviceId }}</span>
    </button>
    <button v-if="editing" type="button" class="nh-tile-del" :aria-label="`Remove ${label} from this area`" @click="emit('remove')">
      <UiIcon name="close" />
    </button>
  </div>
</template>

<style scoped>
  .nh-tile {
    position: relative;
    display: flex;
    align-items: center;
    gap: 0.65rem;
    min-height: var(--nh-tile-h);
    padding: 0.6rem 0.75rem;
    border-radius: var(--nh-radius-l);
    background: var(--nh-surface);
    border: var(--nh-border-w) solid transparent;
    box-shadow: var(--nh-shadow-1);
    transition:
      background var(--nh-motion),
      border-color var(--nh-motion);
    min-width: 0;
  }
  [data-mode='dark'] .nh-tile {
    border-color: var(--nh-border);
  }
  .nh-tile:hover {
    border-color: var(--nh-border-strong);
  }
  .nh-tile-ic {
    display: grid;
    place-items: center;
    width: 2.5rem;
    height: 2.5rem;
    border-radius: 50%;
    flex: none;
    border: 0;
    padding: 0;
    background: var(--nh-surface-2);
    color: var(--nh-kc, var(--nh-text-muted));
    transition:
      background var(--nh-motion),
      box-shadow var(--nh-motion);
  }
  button.nh-tile-ic {
    cursor: pointer;
  }
  button.nh-tile-ic:hover:not(:disabled) {
    box-shadow: 0 0 0 3px var(--nh-accent-soft);
  }
  button.nh-tile-ic:disabled {
    cursor: default;
  }
  .nh-tile-body {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    justify-content: center;
    min-width: 0;
    flex: 1;
    align-self: stretch;
    border: 0;
    background: none;
    color: inherit;
    text-align: left;
    padding: 0;
    cursor: pointer;
  }
  .nh-tile-body:disabled {
    cursor: grab;
  }
  .nh-tile-label {
    font-weight: 600;
    font-size: 0.875rem;
    max-width: 100%;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .nh-tile-value {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    font-size: 0.82rem;
    color: var(--nh-text-muted);
    white-space: nowrap;
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .nh-tile-value .nh-ic {
    width: 0.95rem;
    height: 0.95rem;
  }
  .nh-tile-dev {
    font-size: 0.7rem;
    color: var(--nh-text-muted);
    max-width: 100%;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .nh-tile.is-on {
    background: color-mix(in srgb, var(--nh-kc, var(--nh-accent)) 9%, var(--nh-surface));
  }
  .nh-tile.is-on .nh-tile-ic {
    background: color-mix(in srgb, var(--nh-kc, var(--nh-accent)) 22%, transparent);
  }
  .nh-tile.is-alert {
    border-color: var(--nh-danger);
    background: var(--nh-danger-soft);
  }
  .nh-tile.is-alert .nh-tile-ic {
    background: var(--nh-danger);
    color: #fff;
    animation: nh-pulse 1.2s infinite;
  }
  .nh-tile.is-pending .nh-tile-value {
    color: var(--nh-accent);
  }
  .nh-tile.is-offline,
  .nh-tile.is-disabled,
  .nh-tile.is-missing {
    opacity: 0.75;
  }
  .nh-tile.is-offline .nh-tile-ic,
  .nh-tile.is-disabled .nh-tile-ic,
  .nh-tile.is-missing .nh-tile-ic {
    filter: grayscale(1);
    opacity: 0.6;
  }
  .nh-tile.is-offline .nh-tile-value,
  .nh-tile.is-missing .nh-tile-value {
    color: var(--nh-danger);
  }
  .nh-tile-del {
    position: absolute;
    top: -0.5rem;
    right: -0.5rem;
    width: 1.75rem;
    height: 1.75rem;
    border-radius: 50%;
    border: 2px solid var(--nh-surface);
    background: var(--nh-danger);
    color: #fff;
    display: grid;
    place-items: center;
    cursor: pointer;
    padding: 0;
  }
  .nh-tile-del .nh-ic {
    width: 0.9rem;
    height: 0.9rem;
  }
</style>
