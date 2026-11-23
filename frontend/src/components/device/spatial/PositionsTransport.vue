<script setup lang="ts">
  import { computed } from 'vue';
  import type { SpatialPoint, SpatialSession } from '@/services/spatial-history.service';
  import {
    clock,
    idleSpans,
    movementChapters,
    type PlaybackGap,
    type RecordingNeighbours,
  } from '@/domain/spatial/playback';
  import {
    distanceProgressLabel,
    hasPlayableSpan,
    movementChapterLabel,
    playbackValueLabel,
    recordingPositionLabel,
    transportStatusLabel,
  } from '@/domain/spatial/positions-labels';
  import { positionedSpan } from '@/domain/spatial/positions-strip';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiRange from '@/components/ui/UiRange.vue';

  const props = defineProps<{
    /** Confirmed recording whose points are rendered. */
    session: SpatialSession;
    /** Position among the filtered recordings and their adjacent ids. */
    neighbours: RecordingNeighbours;
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
    /** True when playback advanced to the last recording's end. */
    reachedEnd: boolean;
    /** True while the next recording is loading after a continuous advance. */
    awaitingCrossing: boolean;
    /** A gap longer than the notice threshold that the next crossing will skip. */
    pendingGap: PlaybackGap | null;
    /** Last skipped gap, shown after a continuous advance crossed it. */
    skippedGap: PlaybackGap | null;
  }>();
  const emit = defineEmits<{
    seek: [value: number];
    toggle: [];
    next: [at: number];
    nextRecording: [];
    previousRecording: [];
    'update:speed': [value: number];
    'update:showPath': [value: boolean];
  }>();

  const speedModel = computed<number>({
    get: () => props.speed,
    set: (value) => emit('update:speed', value),
  });
  /*
   * Computed, not a one-time value: the transport stays mounted while the confirmed recording
   * changes, so a plain const would keep the first recording's chapters forever.
   */
  const movements = computed(() =>
    movementChapters(props.points, props.session.from, props.session.to)
  );
  const chapters = computed(() => movements.value.chapters);
  const total = computed(() => movements.value.total);
  const spans = computed(() => idleSpans(props.points, props.session.from, props.session.to));
  const nextEpisode = computed(
    () => chapters.value.find((chapter) => chapter.at > props.time) ?? null
  );

  // Every user-facing sentence comes from `positions-labels`, so the rail, timeline and
  // transport cannot describe the same recording differently.
  const recordingLabel = computed(() =>
    recordingPositionLabel(props.session, props.neighbours)
  );
  const statusText = computed(() =>
    transportStatusLabel({
      skippedGap: props.skippedGap,
      pendingGap: props.pendingGap,
      session: props.session,
      idleForMs: props.idleForMs,
      lastSampleAt: props.lastSampleAt,
      awaitingCrossing: props.awaitingCrossing,
      reachedEnd: props.reachedEnd,
      movementCount: total.value,
      renderedMovements: chapters.value.length,
    })
  );
  const hasSpan = computed(() => hasPlayableSpan(props.session));
  /** Units and wording both live in domain helpers; the component only places the string. */
  const distanceProgress = computed(() =>
    distanceProgressLabel(props.distanceMm, props.totalDistanceMm)
  );
  const valueText = computed(() => playbackValueLabel(props.time, props.session.to));
  const canGoPrevious = computed(() => props.neighbours.previous !== null);
  const canGoNext = computed(() => props.neighbours.next !== null);

  function spanStyle(spanValue: { from: number; to: number }) {
    return positionedSpan(spanValue, { from: props.session.from, to: props.session.to });
  }

  function chapterLabel(chapter: { at: number; slot: number }): string {
    return movementChapterLabel(chapter.at, chapter.slot);
  }
</script>

<template>
  <div class="nh-transport" role="group" aria-label="Playback controls">
    <div class="nh-transport-recordings">
      <UiButton
        size="sm"
        variant="ghost"
        icon="chevronLeft"
        label="Previous recording"
        :disabled="!canGoPrevious"
        @click="emit('previousRecording')" />
      <span class="nh-transport-recording" role="status">{{ recordingLabel }}</span>
      <UiButton
        size="sm"
        variant="ghost"
        icon="chevronRight"
        label="Next recording"
        :disabled="!canGoNext"
        @click="emit('nextRecording')" />
    </div>
    <!--
      Two groups, not seven equal items: the left group is what the operator acts on and what the
      cursor is at (the clock is the loudest number in the card), the right group is settings and
      readouts at one quiet weight.
    -->
    <div class="nh-transport-row">
      <div class="nh-transport-controls">
        <UiButton
          class="nh-transport-play"
          size="sm"
          variant="primary"
          :disabled="!hasPoints"
          @click="emit('toggle')">
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
      </div>
      <div class="nh-transport-metrics">
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
        <span class="nh-transport-distance">{{ distanceProgress }}</span>
      </div>
    </div>
    <!--
      One status line, not three stacked paragraphs. The transport sat as tall as the canvas
      because every state (idle, crossing hint, movement count) took its own row even when empty.
    -->
    <p class="nh-transport-status" aria-live="polite">{{ statusText }}</p>
    <div v-if="hasSpan" class="nh-transport-scrub">
      <div class="nh-transport-marks">
        <span
          v-for="(gap, index) in spans"
          :key="`gap-${index}`"
          class="nh-transport-gap"
          aria-hidden="true"
          :style="spanStyle(gap)" />
        <button
          v-for="chapter in chapters"
          :key="`chapter-${chapter.at}-${chapter.slot}`"
          type="button"
          class="nh-transport-mark"
          :style="{ left: `${chapter.pct}%` }"
          :aria-label="chapterLabel(chapter)"
          :title="chapterLabel(chapter)"
          @click="emit('seek', chapter.at)" />
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
        <span>{{ clock(session.from, { seconds: false }) }}</span>
        <span>{{ clock(session.from + (session.to - session.from) / 2, { seconds: false }) }}</span>
        <span>{{ clock(session.to, { seconds: false }) }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
  .nh-transport {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    padding: 0.45rem 0.75rem;
    border: 1px solid var(--nh-border);
    border-radius: var(--nh-radius-m);
    background: var(--nh-surface);
    box-shadow: var(--nh-shadow-1);
  }
  /*
   * Prev / title / next are one cluster. The label used to stretch, which pushed the next chevron
   * to the card's far edge ~600 px away, where it read as part of the border and operators
   * reported the next-recording control as missing.
   */
  .nh-transport-recordings {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    max-width: 100%;
  }
  .nh-transport-recordings :deep(.nh-btn) {
    flex: none;
  }
  /* A disabled chevron at 0.45 opacity is easy to miss; keep it legible but clearly unavailable. */
  .nh-transport-recordings :deep(.nh-btn:disabled) {
    opacity: 0.6;
  }
  .nh-transport-recording {
    flex: 0 1 auto;
    min-width: 0;
    font-size: 0.9rem;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .nh-transport-row {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem 1rem;
  }
  .nh-transport-controls,
  .nh-transport-metrics {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.4rem 0.75rem;
  }
  .nh-transport-metrics {
    color: var(--nh-text-muted);
    font-size: 0.78rem;
  }
  .nh-transport-row label {
    display: inline-flex;
    align-items: center;
    gap: 0.3rem;
  }
  /* Fixed width stops Play ↔ Pause from shifting the rest of the row. */
  .nh-transport-play {
    min-width: 4.75rem;
    justify-content: center;
  }
  /*
   * The clock leads the row without becoming its monument: one step above body size, tabular so
   * it never jitters. At 1.5rem it was wider than the recording title and out-measured the map.
   */
  .nh-transport-clock {
    font-size: 1.15rem;
    font-weight: 600;
    line-height: 1.2;
    font-variant-numeric: tabular-nums;
    min-width: 5ch;
  }
  /*
   * Reserved height keeps the scrubber from jumping when the sentence changes while the playhead
   * moves; one line is enough for every state because the states are mutually exclusive.
   */
  .nh-transport-status {
    margin: 0;
    min-height: 1rem;
    line-height: 1rem;
    color: var(--nh-text-muted);
    font-size: 0.78rem;
  }
  .nh-transport-scrub {
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
  }
  /* The rail needs a visible extent: in dark themes the track and the card are near-equal. */
  .nh-transport-scrub :deep(.nh-range) {
    height: 1.9rem;
  }
  .nh-transport-scrub :deep(.nh-range-track) {
    box-shadow: inset 0 0 0 1px var(--nh-border);
  }
  .nh-transport-scrub :deep(.nh-range-grip) {
    width: 8px;
    height: 60%;
  }
  .nh-transport-marks {
    position: relative;
    height: 0.85rem;
    margin-bottom: -0.3rem;
    z-index: 1;
  }
  /*
   * Gaps are hatched at a fixed period, so in a dense recording the hatch and the movement marks
   * become two overlapping barcodes. The marks carry the reading; the gap is stated in text.
   */
  .nh-transport-gap {
    position: absolute;
    top: 0;
    bottom: 0;
    border-radius: 2px;
    background: var(--nh-border-strong);
    opacity: 0.4;
  }
  /* A movement is a real control: it can be focused, named and activated, not a decorative tick. */
  .nh-transport-mark {
    position: absolute;
    top: 0;
    width: 0.5rem;
    height: 100%;
    padding: 0;
    border: 0;
    background: none;
    cursor: pointer;
    transform: translateX(-50%);
  }
  .nh-transport-mark::before {
    content: '';
    position: absolute;
    top: 0.15rem;
    bottom: 0.15rem;
    left: 0.2rem;
    width: 0.15rem;
    border-radius: 2px;
    background: var(--nh-text-muted);
  }
  .nh-transport-mark:focus-visible {
    outline: 2px solid var(--nh-focus);
    outline-offset: 1px;
  }
  .nh-transport-labels {
    display: flex;
    justify-content: space-between;
    gap: 0.5rem;
    color: var(--nh-text-muted);
    font-size: 0.78rem;
    font-variant-numeric: tabular-nums;
  }
  @media (max-width: 899px) {
    .nh-transport-recordings :deep(.nh-btn),
    .nh-transport-row :deep(.nh-btn) {
      min-width: 44px;
      min-height: 44px;
    }
    .nh-transport-speed :deep(.nh-input) {
      min-height: 44px;
    }
    /*
     * On a phone the transport already owns most of the width, so the proportional clock would be
     * as tall as the touch targets around it. Step it down once, keep the tabular face.
     */
    .nh-transport-clock {
      font-size: 1rem;
    }
    .nh-transport-row {
      gap: 0.35rem 0.5rem;
    }
  }
</style>
