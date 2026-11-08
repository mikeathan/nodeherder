<script setup lang="ts">
  import { computed } from 'vue';
  import type { SessionSegment } from '@/domain/spatial/playback';
  import { clock, durationLabel } from '@/domain/spatial/playback';

  const props = defineProps<{
    segments: readonly SessionSegment[];
    selectedId: string;
    /** Requested but not yet loaded; shown as pending, never as current. */
    pendingId?: string;
    /** Preset window start, epoch milliseconds. */
    from: number;
    /** Preset window end, epoch milliseconds. */
    to: number;
    /** Playhead, epoch milliseconds. */
    time: number;
    loading?: boolean;
  }>();
  const emit = defineEmits<{ select: [id: string] }>();

  const span = computed(() => Math.max(1, props.to - props.from));
  const showDate = computed(() => span.value > 86_400_000);
  const ticks = computed(() => [props.from, props.from + span.value / 2, props.to]);
  const playhead = computed(() => {
    if (!props.selectedId || props.time <= 0) return null;
    return Math.min(100, Math.max(0, ((props.time - props.from) / span.value) * 100));
  });

  function edgeLabel(at: number) {
    return clock(at, { date: showDate.value, seconds: false });
  }

  function segmentLabel(segment: SessionSegment) {
    const slots = segment.slots.length ? segment.slots.join(', ') : 'none';
    const capped = segment.capped ? ', capped' : '';
    return `${clock(segment.from, { date: showDate.value })} to ${clock(segment.to, {
      date: showDate.value,
    })}, ${durationLabel(segment.to - segment.from)}, ${segment.points} points, slot ${slots}${capped}`;
  }

  function playedPct(segment: SessionSegment) {
    if (segment.id !== props.selectedId || props.time <= segment.from) return 0;
    if (props.time >= segment.to) return 100;
    return ((props.time - segment.from) / Math.max(1, segment.to - segment.from)) * 100;
  }
</script>

<template>
  <div class="nh-coverage">
    <div class="nh-coverage-head">
      <span class="nh-coverage-title">Recordings in range</span>
      <span class="nh-muted">{{ segments.length }} recorded</span>
    </div>
    <div class="nh-coverage-track" role="group" aria-label="Recordings in the selected range">
      <button
        v-for="segment in segments"
        :key="segment.id"
        type="button"
        class="nh-coverage-segment"
        :class="{
          'is-selected': segment.id === selectedId,
          'is-pending': segment.id === pendingId && segment.id !== selectedId,
        }"
        :style="{ left: `${segment.leftPct}%`, width: `${segment.widthPct}%` }"
        :aria-pressed="segment.id === selectedId"
        :aria-busy="segment.id === pendingId || undefined"
        :aria-label="segmentLabel(segment)"
        :title="segmentLabel(segment)"
        :data-session-id="segment.id"
        @click="emit('select', segment.id)">
        <span class="nh-coverage-fill" :style="{ width: `${playedPct(segment)}%` }" />
      </button>
      <span
        v-if="playhead !== null"
        class="nh-coverage-playhead"
        :style="{ left: `${playhead}%` }" />
    </div>
    <div class="nh-coverage-ticks" aria-hidden="true">
      <span v-for="(tick, index) in ticks" :key="index">{{ edgeLabel(tick) }}</span>
    </div>
    <p v-if="!segments.length && loading" class="nh-muted nh-coverage-empty" role="status">
      Loading recordings…
    </p>
    <p v-else-if="!segments.length" class="nh-muted nh-coverage-empty">
      No recordings in this range. Enable “Record history” on the device and let a target move.
    </p>
  </div>
</template>

<style scoped>
  .nh-coverage {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
  }
  .nh-coverage-head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: 0.5rem;
    font-size: 0.8rem;
  }
  .nh-coverage-title {
    font-weight: 600;
  }
  .nh-coverage-track {
    position: relative;
    height: 2.25rem;
    border: 1px solid var(--nh-border);
    border-radius: var(--nh-radius-m);
    background: var(--nh-surface-2);
    overflow: hidden;
  }
  .nh-coverage-segment {
    position: absolute;
    top: 0;
    bottom: 0;
    min-width: 6px;
    padding: 0;
    border: 0;
    border-radius: 2px;
    background: var(--nh-border-strong);
    opacity: 0.6;
    cursor: pointer;
    overflow: hidden;
  }
  .nh-coverage-segment.is-selected {
    opacity: 1;
    background: var(--nh-accent-soft);
    box-shadow: inset 0 0 0 2px var(--nh-accent);
  }
  .nh-coverage-segment.is-pending {
    opacity: 0.8;
    box-shadow: inset 0 0 0 2px var(--nh-border-strong);
  }
  .nh-coverage-segment:focus-visible {
    outline: 2px solid var(--nh-focus);
    outline-offset: -2px;
  }
  .nh-coverage-fill {
    display: block;
    height: 100%;
    background: var(--nh-accent);
    opacity: 0.45;
  }
  .nh-coverage-playhead {
    position: absolute;
    top: 0;
    bottom: 0;
    width: 2px;
    background: var(--nh-text);
    transform: translateX(-1px);
    pointer-events: none;
  }
  .nh-coverage-ticks {
    display: flex;
    justify-content: space-between;
    gap: 0.5rem;
    color: var(--nh-text-muted);
    font-size: 0.72rem;
    font-variant-numeric: tabular-nums;
  }
  .nh-coverage-empty {
    margin: 0;
    font-size: 0.85rem;
  }
</style>
