<script setup lang="ts">
  import { computed, ref } from 'vue';
  import { onBeforeRouteLeave } from 'vue-router';
  import { useEntityDialog } from '@/composables/useEntityDialog';
  import { useHistoryQuery } from '@/composables/useHistoryQuery';
  import { useMediaQuery } from '@/composables/useMediaQuery';
  import {
    binaryStateOf,
    controlOf,
    exposeKind,
    exposeLabel,
    formatExposeValue,
    isWritable,
    measurementExposes,
  } from '@/domain/exposes';
  import type { Expose } from '@/types/device';
  import UiDialog from '@/components/ui/UiDialog.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import EntityAdjustment from '@/components/entity/EntityAdjustment.vue';
  import EntityMeter from '@/components/entity/EntityMeter.vue';
  import EntityReadout from '@/components/entity/EntityReadout.vue';
  import EntityPowerButton from '@/components/entity/EntityPowerButton.vue';
  import ExposeControl from '@/components/entity/ExposeControl.vue';
  import HistoryView from '@/components/chart/history/HistoryView.vue';
  import RelativeTime from '@/components/ui/RelativeTime.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import { exposeIcon } from '@/components/ui/icons';
  const props = defineProps<{ show: boolean; title?: string; id: string; name: string }>();
  const emit = defineEmits<{ close: [value: boolean] }>();
  const model = useEntityDialog(props.id, props.name);
  const { device, expose, adjustment, adjustable, selected, related, status } = model;
  const narrow = useMediaQuery('(max-width: 899px)');
  const now = ref(Date.now());
  const exists = computed(() => !!device.value?.exposes[selected.value]);
  const request = computed(() => {
    if (!exists.value || !props.show) return null;
    return { deviceId: props.id, exposes: [selected.value], from: now.value - 4 * 3600000, to: now.value };
  });
  const { entry, refresh } = useHistoryQuery(request);
  const series = computed(() => entry.value?.series.find((s) => s.name === selected.value));
  const title = computed(() => exposeLabel(selected.value));
  const writable = computed(() => !!adjustment.value && controlOf(adjustment.value) !== 'none');
  const kind = computed(() => exposeKind(selected.value));
  const unknownCurrent = computed(
    () =>
      !adjustment.value ||
      adjustment.value.data === null ||
      adjustment.value.data === undefined ||
      adjustment.value.data === ''
  );
  const current = computed(() =>
    adjustment.value && !unknownCurrent.value ? formatExposeValue(adjustment.value) : 'Unknown'
  );
  const latest = computed(() => {
    const data = series.value?.data ?? [];
    for (let i = data.length - 1; i >= 0; i--) if (data[i].value !== null) return data[i];
    return null;
  });
  const modeControls = computed(() => related.value.filter((control) => isMode(control)));
  const extraControls = computed(() =>
    related.value.filter((control) => !isMode(control) && control.name !== selected.value)
  );
  const alertControl = computed(() =>
    extraControls.value.filter((control) => ['silence', 'test'].includes(control.name))
  );
  const powerControl = computed(() => extraControls.value.find((control) => control.name === 'state'));
  const powerState = computed(() => {
    if (!powerControl.value) return '';
    const state = binaryStateOf(powerControl.value);
    return state === null ? 'Unknown' : state ? 'On' : 'Off';
  });
  const otherExtraControls = computed(() =>
    extraControls.value.filter((control) => control.name !== 'state' && !['silence', 'test'].includes(control.name))
  );
  const hasControls = computed(
    () => modeControls.value.length > 0 || otherExtraControls.value.length > 0 || !!powerControl.value
  );
  const otherReadings = computed(() => {
    if (!device.value || writable.value) return [];
    return measurementExposes(device.value)
      .filter((reading) => reading.name !== selected.value && !isWritable(reading) && hasReading(reading))
      .slice(0, 4);
  });
  const compactHistory = computed(
    () => narrow.value && (writable.value || adjustment.value?.name === 'energy')
  );
  // Geometry is a property of the view, never of an asynchronous history result.
  // Every entity uses the same full-screen phone shell, including offline/empty states.
  const wideDialog = computed(() => adjustable.value);
  function hasReading(reading: Expose) {
    return reading.data !== null && reading.data !== undefined && reading.data !== '';
  }
  function isMode(control: Expose) {
    return ['brightness', 'color_temp'].includes(control.name);
  }
  function controlLabel(name: string) {
    if (name === 'state') return 'Power';
    if (name === 'color_temp') return 'Colour temp';
    return exposeLabel(name);
  }
  function select(name: string) {
    selected.value = name;
  }
  function close() {
    emit('close', false);
  }
  function refreshHistory() {
    now.value = Date.now();
    refresh();
  }
  onBeforeRouteLeave(close);
</script>
<template>
  <UiDialog
    :open="show"
    :title="title"
    :wide="wideDialog"
    fullscreen-mobile
    @close="close">
    <div
      v-if="device && expose && adjustment"
      class="entity-detail"
      :class="[{ 'has-adjustment': adjustable }, `kind-${kind}`]">
      <p class="entity-device">{{ device.friendly_name }}</p>
      <p v-if="status !== 'online'" class="entity-warning" role="status">
        {{ status === 'disabled' ? 'Device disabled' : 'Device offline' }} · controls unavailable
      </p>
      <div class="entity-primary">
        <EntityAdjustment v-if="adjustable" :key="selected" :device="device" :expose="adjustment" />
        <template v-else>
          <template v-if="adjustment.name === 'energy'">
            <EntityMeter :reading="current" />
            <p class="entity-seen">Device seen <RelativeTime :value="device.last_seen" /></p>
          </template>
          <EntityReadout v-else :expose="adjustment" :reading="current" :unknown="unknownCurrent" :seen="device.last_seen" />
          <p v-if="unknownCurrent && latest" class="entity-readout-note">
            Latest recorded {{ formatExposeValue(adjustment, latest.value) }} ·
            <RelativeTime :value="latest.timestamp" />
          </p>
          <p v-else-if="unknownCurrent" class="entity-readout-note">No reading reported.</p>
          <div v-if="writable" class="entity-writable">
            <span class="entity-control-label">{{ controlLabel(adjustment.name) }}</span>
            <ExposeControl :device="device" :expose="adjustment" />
          </div>
        </template>
        <div v-for="action in alertControl" :key="action.name" class="entity-writable">
          <span class="entity-control-label">{{ controlLabel(action.name) }}</span>
          <ExposeControl :device="device" :expose="action" :label="controlLabel(action.name)" />
        </div>
        <div v-if="hasControls" class="entity-controls">
          <button
            v-for="control in modeControls"
            :key="control.name"
            type="button"
            class="entity-mode"
            :aria-label="exposeLabel(control.name)"
            :aria-pressed="selected === control.name"
            @click="select(control.name)">
            <span class="entity-mode-icon"><UiIcon :path="exposeIcon(control.name)" /></span>
            <span class="entity-mode-label">{{ controlLabel(control.name) }}</span>
          </button>
          <EntityPowerButton v-if="powerControl && adjustable" :device="device" :expose="powerControl" />
          <div v-else-if="powerControl" class="entity-extra-control">
            <ExposeControl :device="device" :expose="powerControl" label="Power" />
            <span class="entity-control-label">Power</span>
            <span class="entity-control-state">{{ powerState }}</span>
          </div>
          <div v-for="control in otherExtraControls" :key="control.name" class="entity-extra-control">
            <ExposeControl :device="device" :expose="control" :label="controlLabel(control.name)" />
            <span class="entity-control-label">{{ controlLabel(control.name) }}</span>
          </div>
        </div>
      </div>
      <section class="entity-history" aria-label="Recent history">
        <header>
          <h3>Last 4 hours</h3>
          <UiButton
            variant="ghost"
            icon="refresh"
            label="Refresh history"
            :disabled="entry?.status === 'loading'"
            @click="refreshHistory" />
        </header>
        <p v-if="entry?.status === 'loading'" role="status">Loading history…</p>
        <p v-else-if="entry?.status === 'error'" role="alert">{{ entry.error }}</p>
        <HistoryView
          v-else-if="series"
          :series="series"
          :expose="adjustment"
          :compact="compactHistory" />
        <div v-else class="entity-history-empty">
          <UiIcon name="chart" />
          <p>No recorded history for this entity in the last 4 hours.</p>
          <small>History appears here when the device reports and metrics are recorded.</small>
        </div>
      </section>
      <section v-if="otherReadings.length" class="entity-other" aria-label="Other readings">
        <h3>Other readings</h3>
        <dl>
          <div v-for="reading in otherReadings" :key="reading.name">
            <dt>{{ exposeLabel(reading.name) }}</dt>
            <dd>{{ formatExposeValue(reading) }}</dd>
          </div>
        </dl>
      </section>
    </div>
    <p v-else role="status">This entity is no longer available.</p>
  </UiDialog>
</template>
<style scoped>
  .entity-detail {
    display: grid;
    gap: 1rem;
    min-width: 0;
  }
  .entity-device {
    color: var(--nh-text-muted);
    margin: 0;
    overflow-wrap: anywhere;
  }
  .entity-primary {
    min-width: 0;
  }
  .entity-seen {
    font-size: 0.8rem;
    color: var(--nh-text-muted);
  }
  .entity-readout-note {
    font-size: 0.8rem;
    color: var(--nh-text-muted);
  }
  .entity-warning {
    color: var(--nh-warn);
  }
  .entity-writable {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.75rem;
    padding: 0.5rem 0;
  }
  .entity-controls {
    display: flex;
    flex-wrap: wrap;
    gap: 0.75rem;
    align-items: flex-start;
    justify-content: flex-start;
    padding: 0.75rem 0 0.25rem;
  }
  .entity-controls > * {
    flex: 1 1 0;
    max-width: 6.5rem;
    min-width: 0;
  }
  .entity-control-label {
    font-size: 0.75rem;
    color: var(--nh-text);
    text-align: center;
    overflow-wrap: anywhere;
  }
  .entity-control-state {
    min-height: 1rem;
    font-size: 0.7rem;
    color: var(--nh-text-muted);
  }
  .entity-mode,
  .entity-extra-control {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: flex-start;
    gap: 0.4rem;
    min-width: 0;
    min-height: 44px;
    padding: 0;
    border: 0;
    background: transparent;
    color: var(--nh-text);
    font-size: 0.75rem;
  }
  .entity-mode {
    cursor: pointer;
  }
  .entity-mode-icon {
    display: grid;
    place-items: center;
    width: 3rem;
    height: 3rem;
    border-radius: 50%;
    border: 1px solid var(--nh-border-strong);
    background: var(--nh-surface-2);
    color: var(--nh-text-muted);
    transition:
      background var(--nh-motion),
      border-color var(--nh-motion),
      color var(--nh-motion);
  }
  .entity-mode[aria-pressed='true'] .entity-mode-icon {
    background: var(--nh-accent-soft);
    border-color: var(--nh-accent);
    color: var(--nh-accent);
  }
  .entity-mode[aria-pressed='true'] .entity-mode-label {
    color: var(--nh-accent);
  }
  .entity-mode:focus-visible {
    outline: 3px solid var(--nh-focus);
    outline-offset: 3px;
    border-radius: var(--nh-radius-s);
  }
  .entity-mode-label {
    color: var(--nh-text-muted);
    overflow-wrap: anywhere;
  }
  .entity-extra-control :deep(.nh-xc) {
    min-height: 3rem;
    align-items: center;
  }
  .entity-history {
    min-width: 0;
  }
  .entity-history-empty {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    padding: 0.75rem 0 0.25rem;
    color: var(--nh-text-muted);
  }
  .entity-history-empty :deep(.nh-ic) {
    width: 1.5rem;
    height: 1.5rem;
  }
  .entity-history-empty p,
  .entity-history-empty small {
    margin: 0;
  }
  .entity-history-empty p {
    font-size: 0.85rem;
  }
  .entity-history-empty small {
    font-size: 0.75rem;
  }
  .entity-history > header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  h3 {
    font-size: 0.9rem;
    margin: 0;
    font-weight: 500;
  }
  .entity-other {
    min-width: 0;
    margin-top: 0.25rem;
  }
  .entity-other dl {
    margin: 0.5rem 0 0;
  }
  .entity-other dl > div {
    display: flex;
    justify-content: space-between;
    gap: 1rem;
    padding: 0.6rem 0;
    border-top: 1px solid var(--nh-border);
    font-size: 0.85rem;
  }
  .entity-other dt {
    color: var(--nh-text-muted);
  }
  .entity-other dd {
    margin: 0;
    text-align: right;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
  }
  @media (min-width: 900px) {
    .has-adjustment {
      grid-template-columns: minmax(0, 0.95fr) minmax(0, 1.05fr);
      gap: 1.5rem;
    }
    .has-adjustment .entity-device,
    .has-adjustment .entity-warning {
      grid-column: 1 / -1;
    }
    .has-adjustment .entity-history {
      align-self: start;
    }
    .has-adjustment .entity-other {
      grid-column: 1 / -1;
    }
  }
</style>
