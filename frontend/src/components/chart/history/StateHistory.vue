<script setup lang="ts">
  import { computed, ref } from 'vue';
  import type { Expose } from '@/types/device';
  import type { HistorySeries } from '@/domain/history/types';
  import { stateSegments, stateStats, stateDensity, type StateSegment } from '@/domain/history/state';
  import { binaryOnValue, binaryOffValue, formatExposeValue } from '@/domain/exposes';
  const props = defineProps<{ series: HistorySeries; expose: Expose; compact?: boolean }>();
  const segments = computed(() => stateSegments(props.series, props.expose));
  // Long timelines become a labelled density strip instead of hundreds of thin buttons.
  const DENSE_SEGMENT_LIMIT = 96;
  const dense = computed(() => segments.value.length > DENSE_SEGMENT_LIMIT);
  const density = computed(() => stateDensity(segments.value, props.series.from, props.series.to));
  const stats = computed(() => stateStats(segments.value));
  const selected = ref<StateSegment | null>(null);
  const time = (value: number) => new Date(value).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  const duration = (ms: number) => `${Math.round(ms / 60000)} min`;
  const activeLabel = computed(() => formatExposeValue(props.expose, binaryOnValue(props.expose)));
  const inactiveLabel = computed(() => formatExposeValue(props.expose, binaryOffValue(props.expose)));
  const ticks = computed(() => [props.series.from, (props.series.from + props.series.to) / 2, props.series.to]);
</script>
<template>
  <div class="state-history">
    <template v-if="dense">
      <p class="interval-detail">Activity density · {{ activeLabel }} portion of known time</p>
      <div class="density-strip" role="img" :aria-label="`${activeLabel} duration by time interval`">
        <span
          v-for="(bucket, i) in density"
          :key="i"
          :title="`${time(bucket.from)}–${time(bucket.to)}: ${bucket.percent === null ? 'Unknown' : Math.round(bucket.percent) + '%'}`"
          :style="{ opacity: bucket.percent === null ? 0.15 : 0.25 + (bucket.percent / 100) * 0.75 }" />
      </div>
    </template>
    <div v-else class="state-strip" role="group" aria-label="State history intervals">
      <button
        v-for="(segment, i) in segments"
        :key="i"
        type="button"
        :class="{
          'is-active': segment.active === true,
          'is-inactive': segment.active === false,
          'is-unknown': segment.active === null,
        }"
        :style="{ width: `${((segment.to - segment.from) / (series.to - series.from)) * 100}%` }"
        :aria-label="`${segment.label}, ${time(segment.from)} to ${time(segment.to)}`"
        @click="selected = segment" />
    </div>
    <div class="state-times">
      <span v-for="tick in ticks" :key="tick">{{ time(tick) }}</span>
    </div>
    <div class="state-legend">
      <span class="active">{{ activeLabel }}</span
      ><span>{{ inactiveLabel }}</span
      ><span class="unknown">Unknown</span>
    </div>
    <p v-if="selected" class="interval-detail" role="status">
      {{ selected.label }} · {{ time(selected.from) }}–{{ time(selected.to) }} ·
      {{ duration(selected.to - selected.from) }}
    </p>
    <dl v-if="!compact" class="history-summary">
      <div>
        <dt>{{ activeLabel }} duration</dt>
        <dd>{{ stats.duration ? duration(stats.active) : 'Unknown' }}</dd>
      </div>
      <div>
        <dt>Transitions</dt>
        <dd>{{ stats.transitions }}</dd>
      </div>
      <div>
        <dt>Known coverage</dt>
        <dd>{{ Math.round((stats.duration / (series.to - series.from)) * 100) }} %</dd>
      </div>
    </dl>
  </div>
</template>
<style scoped>
  .density-strip {
    display: flex;
    height: 2rem;
    overflow: hidden;
    border-radius: var(--nh-radius-s);
  }
  .density-strip span {
    flex: 1;
    background: var(--history-color);
  }
  .state-strip {
    display: flex;
    height: 2rem;
    overflow: hidden;
    border-radius: var(--nh-radius-s);
  }
  .state-strip button {
    flex: none;
    min-width: 0;
    padding: 0;
    border: 0;
    cursor: pointer;
    background: var(--nh-text-muted);
  }
  .state-strip button:focus-visible {
    outline: 3px solid var(--nh-focus);
    outline-offset: -3px;
  }
  .state-strip .is-active {
    background: var(--history-color);
  }
  .state-strip .is-unknown {
    background: repeating-linear-gradient(
      45deg,
      var(--nh-surface-2),
      var(--nh-surface-2) 4px,
      var(--nh-border) 4px,
      var(--nh-border) 8px
    );
  }
  .state-times {
    display: flex;
    justify-content: space-between;
    color: var(--nh-text-muted);
    font-size: 0.75rem;
    margin-top: 0.5rem;
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
    background: var(--nh-text-muted);
    margin-right: 0.4rem;
  }
  .state-legend .active::before {
    background: var(--history-color);
  }
  .state-legend .unknown::before {
    background: var(--nh-surface-2);
    border: 1px solid var(--nh-border-strong);
  }
  .interval-detail {
    font-size: 0.8rem;
    color: var(--nh-text-muted);
  }
</style>
