<script setup lang="ts">
  import { computed, ref } from 'vue';
  import type { Expose } from '@/types/device';
  import { stateDensity, type StateSegment } from '@/domain/history/state';
  import { binaryOnValue, exposePalette, formatExposeValue } from '@/domain/exposes';
  import {
    formatHistoryClock,
    formatHistoryMinutes,
    formatHistorySharePercent,
  } from '@/domain/history/format';

  const props = defineProps<{ segments: StateSegment[]; expose: Expose; from: number; to: number }>();

  const palette = computed(() => exposePalette(props.expose));
  const activeLabel = computed(() => formatExposeValue(props.expose, binaryOnValue(props.expose)));
  /** Long compact timelines become a density strip instead of thin, unreadable buttons. */
  const dense = computed(() => props.segments.length > 96);
  const density = computed(() => stateDensity(props.segments, props.from, props.to));
  const selected = ref<StateSegment | null>(null);
  const ticks = computed(() => [props.from, (props.from + props.to) / 2, props.to]);
  const densityTitle = (bucket: { from: number; to: number; percent: number | null }) => {
    const share = formatHistorySharePercent(bucket.percent);
    return `${formatHistoryClock(bucket.from)}–${formatHistoryClock(bucket.to)}: ${share ?? 'Unknown'}`;
  };
</script>

<template>
  <div class="state-compact" :style="{ '--state-active': palette.active, '--state-inactive': palette.inactive }">
    <p v-if="dense" class="interval-detail">Activity density · {{ activeLabel }} portion of known time</p>
    <div v-if="dense" class="density-strip" role="img" :aria-label="`${activeLabel} duration by time interval`">
      <span
        v-for="(bucket, index) in density"
        :key="index"
        :title="densityTitle(bucket)"
        :style="{ opacity: bucket.percent === null ? 0.15 : 0.25 + (bucket.percent / 100) * 0.75 }" />
    </div>
    <div v-else class="state-strip" role="group" aria-label="State history intervals">
      <button
        v-for="(segment, index) in segments"
        :key="index"
        type="button"
        :class="{
          'is-active': segment.active === true,
          'is-inactive': segment.active === false,
          'is-unknown': segment.active === null,
        }"
        :style="{ width: `${((segment.to - segment.from) / (to - from)) * 100}%` }"
        :aria-label="`${segment.label}, ${formatHistoryClock(segment.from)} to ${formatHistoryClock(segment.to)}`"
        @click="selected = segment" />
    </div>
    <div class="state-times">
      <span v-for="tick in ticks" :key="tick">{{ formatHistoryClock(tick) }}</span>
    </div>
    <p v-if="selected" class="interval-detail" role="status">
      {{ selected.label }} · {{ formatHistoryClock(selected.from) }}–{{ formatHistoryClock(selected.to) }} ·
      {{ formatHistoryMinutes(selected.to - selected.from) }}
    </p>
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
    background: var(--state-active);
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
    background: var(--state-inactive);
  }
  .state-strip button:focus-visible {
    outline: 3px solid var(--nh-focus);
    outline-offset: -3px;
  }
  .state-strip .is-active {
    background: var(--state-active);
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
  .interval-detail {
    font-size: 0.8rem;
    color: var(--nh-text-muted);
  }
</style>
