<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import { findDevice } from '@/composables/useHub';
  import { useHistoryQuery } from '@/composables/useHistoryQuery';
  import { historyRange, validHistoryRange } from '@/domain/history/range';
  import { exposeLabel } from '@/domain/exposes';
  import HistoryView from '@/components/chart/history/HistoryView.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  const props = defineProps<{ id: string }>();
  const device = computed(() => findDevice(props.id));
  const names = computed(() => Object.keys(device.value?.exposes ?? {}).sort());
  const selected = ref('');
  const hours = ref('24');
  const end = ref(Date.now());
  const customFrom = ref('');
  const customTo = ref('');
  const periods = [
    { value: '1', label: '1 hour' },
    { value: '6', label: '6 hours' },
    { value: '24', label: '24 hours' },
    { value: '168', label: '7 days' },
    { value: '720', label: '30 days' },
    { value: 'today', label: 'Today' },
    { value: 'custom', label: 'Custom' },
  ];
  watch(
    names,
    (values) => {
      if (!values.includes(selected.value)) selected.value = values[0] ?? '';
    },
    { immediate: true }
  );
  const range = computed(() => historyRange(hours.value, end.value, customFrom.value, customTo.value));
  const from = computed(() => range.value.from);
  const to = computed(() => range.value.to);
  const valid = computed(() => validHistoryRange(range.value));
  const request = computed(() =>
    valid.value && selected.value
      ? { deviceId: props.id, exposes: [selected.value], from: from.value, to: to.value }
      : null
  );
  const { entry, refresh } = useHistoryQuery(request);
  const series = computed(() => entry.value?.series.find((s) => s.name === selected.value));
  const expose = computed(() => device.value?.exposes[selected.value]);
  function reload() {
    end.value = Date.now();
    refresh();
  }
</script>
<template>
  <section class="device-history" aria-label="Device history">
    <div class="history-toolbar">
      <label
        >Entity<select aria-label="Entity" v-model="selected" class="nh-input">
          <option v-for="name in names" :key="name" :value="name">{{ exposeLabel(name) }}</option>
        </select></label
      >
      <label
        >Period<select aria-label="Period" v-model="hours" class="nh-input">
          <option v-for="period in periods" :key="period.value" :value="period.value">{{ period.label }}</option>
        </select></label
      >
      <UiButton
        variant="ghost"
        icon="refresh"
        label="Refresh history"
        :disabled="entry?.status === 'loading'"
        @click="reload" />
    </div>
    <div v-if="hours === 'custom'" class="history-toolbar">
      <label>From<input v-model="customFrom" class="nh-input" type="datetime-local" /></label
      ><label>To<input v-model="customTo" class="nh-input" type="datetime-local" /></label>
    </div>
    <p class="history-zone">
      {{ Intl.DateTimeFormat().resolvedOptions().timeZone
      }}<template v-if="valid"> · {{ new Date(from).toLocaleString() }} – {{ new Date(to).toLocaleString() }}</template>
    </p>
    <p v-if="!valid" role="status">Choose a valid period of at most 31 days.</p>
    <p v-else-if="entry?.status === 'loading'" role="status">Loading history…</p>
    <p v-else-if="entry?.status === 'error'" role="alert">{{ entry.error }}</p>
    <HistoryView v-else-if="series && expose" :series="series" :expose="expose" />
    <p v-else>No recorded history for this entity.</p>
  </section>
</template>
<style scoped>
  .device-history {
    max-width: 64rem;
    margin: auto;
    padding: 0.5rem 0;
  }
  .history-toolbar {
    display: flex;
    flex-wrap: wrap;
    align-items: end;
    gap: 0.75rem;
  }
  label {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
    flex: 1 1 10rem;
    font-size: 0.8rem;
  }
  select,
  input {
    width: 100%;
    min-height: 44px;
  }
  .history-zone {
    font-size: 0.75rem;
    color: var(--nh-text-muted);
    overflow-wrap: anywhere;
    margin: 1rem 0;
  }
</style>
