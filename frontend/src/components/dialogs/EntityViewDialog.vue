<script setup lang="ts">
  import { computed, ref } from 'vue';
  import { onBeforeRouteLeave } from 'vue-router';
  import { useEntityDialog } from '@/composables/useEntityDialog';
  import { useHistoryQuery } from '@/composables/useHistoryQuery';
  import { exposeLabel, formatExposeValue, controlOf, exposeKind } from '@/domain/exposes';
  import UiDialog from '@/components/ui/UiDialog.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import EntityAdjustment from '@/components/entity/EntityAdjustment.vue';
  import ExposeControl from '@/components/entity/ExposeControl.vue';
  import HistoryView from '@/components/chart/history/HistoryView.vue';
  import RelativeTime from '@/components/ui/RelativeTime.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import { exposeIcon } from '@/components/ui/icons';
  const props = defineProps<{ show: boolean; title?: string; id: string; name: string }>();
  const emit = defineEmits<{ close: [value: boolean] }>();
  const model = useEntityDialog(props.id, props.name);
  const { device, expose, adjustment, adjustable, selected, related, status } = model;
  const now = ref(Date.now());
  const exists = computed(() => !!device.value?.exposes[selected.value]);
  const request = computed(() => {
    if (!exists.value || !props.show) return null;
    return { deviceId: props.id, exposes: [selected.value], from: now.value - 4 * 3600000, to: now.value };
  });
  const { entry, refresh } = useHistoryQuery(request);
  const series = computed(() => entry.value?.series.find((s) => s.name === selected.value));
  const title = computed(() => exposeLabel(selected.value));
  const writable = computed(() => !!expose.value && controlOf(expose.value) !== 'none');
  const kind = computed(() => exposeKind(selected.value));
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
  <UiDialog :open="show" :title="title" wide fullscreen-mobile @close="close">
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
        <template v-else
          ><div class="entity-current">{{ formatExposeValue(adjustment) }}</div>
          <p v-if="adjustment.name === 'energy'" class="entity-seen">Meter total</p>
          <p class="entity-seen">Device seen <RelativeTime :value="device.last_seen" /></p>
          <ExposeControl v-if="writable" :device="device" :expose="expose"
        /></template>
        <div v-if="related.length > 1" class="entity-controls">
          <div v-for="control in related" :key="control.name">
            <button
              v-if="['brightness', 'color_temp'].includes(control.name)"
              type="button"
              class="entity-tab"
              :aria-label="exposeLabel(control.name)"
              :aria-pressed="selected === control.name"
              @click="selected = control.name">
              <UiIcon :path="exposeIcon(control.name)" />{{
                control.name === 'color_temp' ? 'Colour temp' : exposeLabel(control.name)
              }}
            </button>
            <template v-else-if="control.name !== selected"
              ><span>{{ control.name === 'state' ? 'Power' : exposeLabel(control.name) }}</span
              ><ExposeControl
                :device="device"
                :expose="control"
                :label="control.name === 'state' ? 'Power' : exposeLabel(control.name)"
            /></template>
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
        <HistoryView v-else-if="series" :series="series" :expose="adjustment" :compact="adjustable" />
        <p v-else>No recorded history for this entity.</p>
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
  .entity-current {
    font-size: clamp(2.5rem, 6vw, 3.5rem);
    font-weight: 600;
    line-height: 1.2;
    letter-spacing: -0.04em;
    padding: 1rem 0 0.5rem;
  }
  .entity-seen {
    font-size: 0.8rem;
    color: var(--nh-text-muted);
  }
  .entity-warning {
    color: var(--nh-warn);
  }
  .entity-history {
    min-width: 0;
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
  .entity-controls {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 0.5rem;
    padding: 0.5rem 0;
  }
  .entity-controls > div {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    align-items: center;
    justify-content: center;
    font-size: 0.75rem;
  }
  .entity-tab {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.4rem;
    padding: 0.7rem 0.4rem;
    width: 100%;
    min-height: 44px;
    border: 1px solid var(--nh-border-strong);
    border-radius: var(--nh-radius-l);
    background: var(--nh-surface-2);
    color: var(--nh-text);
    cursor: pointer;
  }
  .entity-tab[aria-pressed='true'] {
    color: var(--nh-accent);
    border-color: var(--nh-accent);
  }
  @media (min-width: 900px) {
    .has-adjustment {
      grid-template-columns: 1fr 1fr;
      gap: 1.5rem;
    }
    .has-adjustment .entity-device,
    .has-adjustment .entity-warning {
      grid-column: 1 / -1;
    }
    .has-adjustment .entity-history {
      align-self: center;
    }
  }
</style>
