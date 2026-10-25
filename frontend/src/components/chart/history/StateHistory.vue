<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import type { Expose } from '@/types/device';
  import type { HistorySeries } from '@/domain/history/types';
  import { recentActivityWindow, stateSegments, stateStats } from '@/domain/history/state';
  import { binaryOnValue, binaryOffValue, exposePalette, formatExposeValue } from '@/domain/exposes';
  import { formatHistoryCoverage, formatHistoryMinutes, formatHistoryMoment, formatHistorySpan } from '@/domain/history/format';
  import { stateFocusIsUseful } from './stateLayout';
  import StateHistoryChart from './StateHistoryChart.vue';
  import StateHistoryCompact from './StateHistoryCompact.vue';

  const props = defineProps<{ series: HistorySeries; expose: Expose; compact?: boolean }>();

  const segments = computed(() => stateSegments(props.series, props.expose));
  const stats = computed(() => stateStats(segments.value));
  const palette = computed(() => exposePalette(props.expose));
  const activeLabel = computed(() => formatExposeValue(props.expose, binaryOnValue(props.expose)));
  const inactiveLabel = computed(() => formatExposeValue(props.expose, binaryOffValue(props.expose)));

  /** Local display window for "Zoom to activity"; never changes the loaded series. */
  const focusRange = ref<{ from: number; to: number } | null>(null);
  const plotWidth = ref(0);
  const announcement = ref('');
  const viewFrom = computed(() => focusRange.value?.from ?? props.series.from);
  const viewTo = computed(() => focusRange.value?.to ?? props.series.to);
  const activity = computed(() => recentActivityWindow(segments.value, props.series.from, props.series.to));
  const canFocus = computed(
    () =>
      !focusRange.value &&
      !!activity.value &&
      stateFocusIsUseful(
        activity.value.clusterFrom,
        activity.value.clusterTo,
        props.series.from,
        props.series.to,
        plotWidth.value
      )
  );
  const isFocused = computed(() => !!focusRange.value);

  const fullRangeLabel = computed(
    () => `${formatHistoryMoment(props.series.from)} – ${formatHistoryMoment(props.series.to)}`
  );
  const focusRangeLabel = computed(
    () => `${formatHistoryMoment(viewFrom.value)} – ${formatHistoryMoment(viewTo.value)}`
  );
  const activityDetail = computed(() =>
    activity.value
      ? `${activity.value.transitionCount} transitions · ` +
        `${formatHistoryMinutes(activity.value.clusterTo - activity.value.clusterFrom)}`
      : ''
  );
  const coverage = computed(() => {
    const span = props.series.to - props.series.from;
    return span > 0 ? (stats.value.duration / span) * 100 : 0;
  });
  const coverageNote = computed(() =>
    coverage.value < 50
      ? `${formatHistorySpan(stats.value.duration)} recorded · ` +
        `${formatHistorySpan(props.series.to - props.series.from - stats.value.duration)} no data`
      : ''
  );
  const summaryDuration = (ms: number) => (stats.value.duration ? formatHistoryMinutes(ms) : 'Unknown');

  function focusActivity() {
    if (!activity.value) return;
    focusRange.value = { from: activity.value.from, to: activity.value.to };
    announcement.value = `Activity detail shown: ${focusRangeLabel.value}. Selected range unchanged: ${fullRangeLabel.value}.`;
  }
  function resetFocus() {
    if (!focusRange.value) return;
    focusRange.value = null;
    announcement.value = `Full range shown: ${fullRangeLabel.value}.`;
  }
  // A new query replaces the series; focus must never leak into another period.
  watch(
    () => [props.series.name, props.series.from, props.series.to],
    () => { focusRange.value = null; announcement.value = ''; },
    { flush: 'sync' }
  );
</script>

<template>
  <div
    class="state-history"
    :style="{ '--state-active': palette.active, '--state-inactive': palette.inactive }">
    <p class="sr-only" role="status" aria-live="polite">{{ announcement }}</p>
    <div v-if="!compact" class="state-toolbar">
      <p class="state-range">
        <strong>{{ isFocused ? 'Activity detail' : 'Full range' }}</strong>
        <span>{{ isFocused ? focusRangeLabel : fullRangeLabel }}</span>
        <span v-if="isFocused" class="state-range-full">Selected range: {{ fullRangeLabel }}</span>
      </p>
      <button v-if="isFocused" type="button" class="nh-btn" @click="resetFocus">Back to full range</button>
      <button
        v-else-if="canFocus"
        type="button"
        class="nh-btn"
        :aria-label="`Zoom to activity: ${activityDetail}`"
        @click="focusActivity">
        <span>Zoom to activity</span>
        <small>{{ activityDetail }}</small>
      </button>
    </div>
    <p v-if="!compact && coverageNote" class="state-coverage">{{ coverageNote }}</p>
    <StateHistoryChart
      v-if="!compact"
      :segments="segments"
      :expose="expose"
      :from="series.from"
      :to="series.to"
      :view-from="viewFrom"
      :view-to="viewTo"
      @plot-width="plotWidth = $event" />
    <StateHistoryCompact v-else :segments="segments" :expose="expose" :from="series.from" :to="series.to" />
    <div class="state-legend">
      <span class="active">{{ activeLabel }}</span>
      <span>{{ inactiveLabel }}</span>
      <span class="unknown">Unknown</span>
    </div>
    <dl v-if="!compact" class="history-summary">
      <div>
        <dt>{{ activeLabel }} duration</dt>
        <dd>{{ summaryDuration(stats.active) }}</dd>
      </div>
      <div>
        <dt>Transitions</dt>
        <dd>{{ stats.transitions }}</dd>
      </div>
      <div>
        <dt>Known coverage</dt>
        <dd>{{ formatHistoryCoverage(stats.duration, series.to - series.from) }}</dd>
      </div>
    </dl>
  </div>
</template>

<style scoped>
  .state-toolbar {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem 1rem;
    margin-bottom: 0.35rem;
  }
  .state-range {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 0.25rem 0.6rem;
    margin: 0;
    color: var(--nh-text-muted);
    font-size: 0.78rem;
  }
  .state-range strong {
    color: var(--nh-text);
  }
  .state-range-full {
    flex-basis: 100%;
    font-size: 0.72rem;
  }
  .state-toolbar .nh-btn {
    display: inline-flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 0.1rem;
    min-height: 44px;
    text-align: left;
  }
  .state-toolbar .nh-btn small {
    color: var(--nh-text-muted);
    font-size: 0.7rem;
  }
  .state-coverage {
    margin: 0 0 0.45rem;
    color: var(--nh-text-muted);
    font-size: 0.75rem;
  }
  .state-legend {
    display: flex;
    flex-wrap: wrap;
    gap: 1rem;
    font-size: 0.75rem;
    margin: 1rem 0;
  }
  .state-legend span::before {
    content: '';
    display: inline-block;
    width: 0.6rem;
    height: 0.6rem;
    border-radius: 50%;
    background: var(--state-inactive);
    margin-right: 0.4rem;
  }
  .state-legend .active::before {
    background: var(--state-active);
  }
  .state-legend .unknown::before {
    background: var(--nh-surface-2);
    border: 1px solid var(--nh-border-strong);
  }
</style>
