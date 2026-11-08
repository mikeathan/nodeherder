<script setup lang="ts">
  import { computed } from 'vue';
  import type { SpatialPoint, SpatialSession } from '@/services/spatial-history.service';
  import {
    clock,
    durationLabel,
    idleSpans,
    movementEpisodes,
    nextMovement,
  } from '@/domain/spatial/playback';
  import { formatDistance } from '@/utils/units';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiRange from '@/components/ui/UiRange.vue';

  const props = defineProps<{
    session: SpatialSession;
    time: number;
    playing: boolean;
    speed: number;
    showPath: boolean;
    points: readonly SpatialPoint[];
    /** Distance covered up to the playhead, in millimetres. */
    distanceMm: number;
    /** Whole-session distance, in millimetres. */
    totalDistanceMm: number;
    /** Elapsed time without a recorded sample, or 0 when movement is current. */
    idleForMs: number;
    lastSampleAt: number;
    hasPoints: boolean;
  }>();
  const emit = defineEmits<{
    seek: [value: number];
    toggle: [];
    next: [at: number];
    'update:speed': [value: number];
    'update:showPath': [value: boolean];
  }>();

  const speedModel = computed<number>({
    get: () => props.speed,
    set: (value) => emit('update:speed', value),
  });
  const episodes = computed(() => movementEpisodes(props.points));
  const spans = computed(() => idleSpans(props.points, props.session.from, props.session.to));
  // Marks are decorative; cap them so a dense session cannot create thousands of nodes.
  const MAX_MARKS = 200;
  function sampled<T>(items: readonly T[]): T[] {
    if (items.length <= MAX_MARKS) return [...items];
    const step = Math.ceil(items.length / MAX_MARKS);
    return items.filter((_, index) => index % step === 0);
  }
  const episodeMarks = computed(() => sampled(episodes.value));
  const gapMarks = computed(() => sampled(spans.value));
  const nextEpisode = computed(() => nextMovement(episodes.value, props.time));
  const span = computed(() => Math.max(1, props.session.to - props.session.from));
  const distanceLabel = computed(() => formatDistance(props.distanceMm / 1000));
  const totalLabel = computed(() => formatDistance(props.totalDistanceMm / 1000));
  const valueText = computed(
    () => `${clock(props.time)} of ${clock(props.session.to)}`
  );
  const idleText = computed(() =>
    props.idleForMs > 0
      ? `No movement recorded for ${durationLabel(props.idleForMs)} · last sample ${clock(props.lastSampleAt)}`
      : ''
  );

  function atPct(at: number) {
    return Math.min(100, Math.max(0, ((at - props.session.from) / span.value) * 100));
  }

  function spanStyle(spanValue: { from: number; to: number }) {
    const left = atPct(spanValue.from);
    const right = atPct(spanValue.to);
    return { left: `${left}%`, width: `${Math.max(0, right - left)}%` };
  }
</script>

<template>
  <div class="nh-transport" role="group" aria-label="Playback controls">
    <div class="nh-transport-row">
      <UiButton class="nh-transport-play" size="sm" :disabled="!hasPoints" @click="emit('toggle')">
        {{ playing ? 'Pause' : 'Play' }}
      </UiButton>
      <UiButton
        size="sm"
        variant="ghost"
        :disabled="!hasPoints || !nextEpisode"
        title="Seek to the next recorded movement"
        @click="nextEpisode && emit('next', nextEpisode.at)">
        Next movement
      </UiButton>
      <span class="nh-transport-clock">{{ clock(time) }}</span>
      <label class="nh-transport-speed">Speed
        <select v-model.number="speedModel" class="nh-input">
          <option :value="1">1×</option>
          <option :value="2">2×</option>
          <option :value="4">4×</option>
        </select>
      </label>
      <label class="nh-transport-path">
        <input
          type="checkbox"
          :checked="showPath"
          @change="emit('update:showPath', ($event.target as HTMLInputElement).checked)" />
        Path
      </label>
      <span class="nh-muted">{{ distanceLabel }} travelled</span>
      <span v-if="totalDistanceMm > distanceMm" class="nh-muted">of {{ totalLabel }}</span>
    </div>
    <p class="nh-transport-idle" aria-live="polite">{{ idleText }}</p>
    <div class="nh-transport-scrub">
      <div class="nh-transport-marks" aria-hidden="true">
        <span
          v-for="(gap, index) in gapMarks"
          :key="`gap-${index}`"
          class="nh-transport-gap"
          :style="spanStyle(gap)" />
        <span
          v-for="episode in episodeMarks"
          :key="`episode-${episode.at}-${episode.slot}`"
          class="nh-transport-mark"
          :style="{ left: `${atPct(episode.at)}%` }" />
      </div>
      <UiRange
        :value="time"
        :min="session.from"
        :max="session.to"
        :step="200"
        label="Playback position"
        :value-text="valueText"
        :disabled="!hasPoints"
        @preview="emit('seek', $event)"
        @commit="emit('seek', $event)" />
      <div class="nh-transport-labels" aria-hidden="true">
        <span>{{ clock(session.from) }}</span>
        <span>{{ durationLabel(session.to - session.from) }}</span>
        <span>{{ clock(session.to) }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
  .nh-transport {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    padding: 0.5rem 0.65rem;
    border: 1px solid var(--nh-border);
    border-radius: var(--nh-radius-m);
    background: var(--nh-surface);
    box-shadow: var(--nh-shadow-1);
  }
  .nh-transport-row {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.4rem 0.75rem;
  }
  .nh-transport-row label {
    display: inline-flex;
    align-items: center;
    gap: 0.3rem;
  }
  /* Fixed width stops Play ↔ Pause from shifting the rest of the row. */
  .nh-transport-play {
    min-width: 4.5rem;
    justify-content: center;
  }
  .nh-transport-clock {
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    min-width: 8.5rem;
  }
  /* Always present so the idle notice can appear without reflowing the transport. */
  .nh-transport-idle {
    margin: 0;
    min-height: 1rem;
    line-height: 1rem;
    color: var(--nh-text-muted);
    font-size: 0.8rem;
  }
  .nh-transport-scrub {
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
  }
  .nh-transport-marks {
    position: relative;
    height: 0.45rem;
  }
  .nh-transport-gap {
    position: absolute;
    top: 0;
    bottom: 0;
    border-radius: 2px;
    background: repeating-linear-gradient(
      135deg,
      var(--nh-border-strong) 0,
      var(--nh-border-strong) 2px,
      transparent 2px,
      transparent 6px
    );
    opacity: 0.55;
  }
  .nh-transport-mark {
    position: absolute;
    top: 0;
    width: 2px;
    height: 100%;
    background: var(--nh-text-muted);
    transform: translateX(-1px);
  }
  .nh-transport-labels {
    display: flex;
    justify-content: space-between;
    gap: 0.5rem;
    color: var(--nh-text-muted);
    font-size: 0.72rem;
    font-variant-numeric: tabular-nums;
  }
  @media (max-width: 899px) {
    .nh-transport-row :deep(.nh-btn) {
      min-width: 44px;
      min-height: 44px;
    }
    .nh-transport-speed :deep(.nh-input) {
      min-height: 44px;
    }
  }
</style>
