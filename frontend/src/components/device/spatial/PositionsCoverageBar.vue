<script setup lang="ts">
  import { computed, ref } from 'vue';
  import type { SessionSegment } from '@/domain/spatial/playback';
  import { clock } from '@/domain/spatial/playback';
  import {
    countLabel,
    fallbackMarkLabel,
    movementMarkLabel,
    rangeHeadlineLabel,
    RANGE_EMPTY_COPY,
    RECORDINGS_LOADING_COPY,
    segmentLabel,
    timeLabelOptions,
  } from '@/domain/spatial/positions-labels';
  import {
    marksPerSegment,
    nearestPositionAt,
    renderSegments,
    sampleMarks,
    totalSampleCount,
  } from '@/domain/spatial/positions-strip';
  import { useElementSize } from '@/composables/positions/useElementSize';

  const props = defineProps<{
    segments: readonly SessionSegment[];
    selectedId: string;
    /** Requested but not yet loaded; shown as pending, never as current. */
    pendingId?: string;
    /** Filter start, epoch milliseconds. */
    from: number;
    /** Filter end, epoch milliseconds. */
    to: number;
    loading?: boolean;
  }>();
  const emit = defineEmits<{
    /** Select a recording and play from a recorded moment inside it, in one request. */
    'play-from': [payload: { id: string; at: number }];
  }>();

  const track = ref<HTMLElement | null>(null);
  // Observed size lives in one place, shared with the viewer; the bar only needs the width.
  const { width: trackWidth } = useElementSize(track);
  const activeMarkKey = ref('');

  const markKey = (segmentId: string, at: number | null) =>
    at === null ? `${segmentId}:fallback` : `${segmentId}:${at}`;

  const span = computed(() => Math.max(1, props.to - props.from));
  const showDate = computed(() => timeLabelOptions(span.value).date);
  const ticks = computed(() => [props.from, props.from + span.value / 2, props.to]);
  const headline = computed(() => rangeHeadlineLabel(props.segments.length, props.from, props.to));
  const totalSamples = computed(() => totalSampleCount(props.segments));
  /*
   * One global mark budget, split evenly across the recordings. Without it a 500-recording filter
   * could put thousands of focus stops ahead of the transport. A segment with no marks still gets
   * one fallback control when it has a share, so a recording without `samples` (older hub) stays
   * reachable by keyboard; when the budget is too small for every segment, only the selected and
   * pending recordings keep a focus stop and the rail remains the full selection surface.
   */
  const rendered = computed(() => {
    const layout = renderSegments(props.segments, trackWidth.value);
    const share = marksPerSegment(layout.length);
    return layout.map((entry) => {
      const marks = sampleMarks(entry.segment.activity, share);
      const fallback =
        marks.length === 0 &&
        (share > 0 ||
          entry.segment.id === props.selectedId ||
          entry.segment.id === props.pendingId);
      return { ...entry, marks, fallback };
    });
  });
  const shownMarkCount = computed(() =>
    rendered.value.reduce((total, entry) => total + entry.marks.length, 0)
  );

  /*
   * Roving tabindex: the timeline is one tab stop, not one per mark. A 24 h filter can expose
   * 120 marks, and making every one a Tab stop puts that many presses between the page and the
   * transport. Arrow keys walk the marks inside the widget; Enter/Space activates.
   */
  const firstMarkKey = computed(() => {
    for (const entry of rendered.value) {
      if (entry.marks.length) return markKey(entry.segment.id, entry.marks[0].at);
      if (entry.fallback) return markKey(entry.segment.id, null);
    }
    return '';
  });
  const markKeys = computed(() => {
    const keys = new Set<string>();
    for (const entry of rendered.value) {
      for (const mark of entry.marks) keys.add(markKey(entry.segment.id, mark.at));
      if (entry.fallback) keys.add(markKey(entry.segment.id, null));
    }
    return keys;
  });
  const tabStopKey = computed(() =>
    markKeys.value.has(activeMarkKey.value) ? activeMarkKey.value : firstMarkKey.value
  );

  /**
   * Insets a mark so its 1.1 rem box stays inside the track: a tick at 0 % or 100 % would
   * otherwise be half-clipped by the track's overflow. The rem term shifts 0→+0.55 rem and
   * 100 %→−0.55 rem while the percentage keeps the spacing monotonic.
   */
  function markLeft(pct: number): string {
    const inset = (0.55 * (1 - (2 * pct) / 100)).toFixed(3);
    return `calc(${pct}% + ${inset}rem)`;
  }

  function onMarkKeydown(event: KeyboardEvent, key: string): void {
    if (event.key !== 'ArrowRight' && event.key !== 'ArrowLeft') return;
    const buttons = Array.from(
      track.value?.querySelectorAll<HTMLButtonElement>('.nh-coverage-mark') ?? []
    );
    const index = buttons.findIndex((button) => button.dataset.key === key);
    if (index < 0) return;
    const next = buttons[event.key === 'ArrowRight' ? index + 1 : index - 1];
    if (!next) return;
    event.preventDefault();
    activeMarkKey.value = next.dataset.key ?? '';
    next.focus();
  }

  function edgeLabel(at: number): string {
    return clock(at, { date: showDate.value, seconds: false });
  }

  function label(segment: SessionSegment): string {
    return segmentLabel(segment, showDate.value);
  }

  function markLabel(segment: SessionSegment, at: number): string {
    return movementMarkLabel(segment, at, showDate.value);
  }

  function fallbackLabel(segment: SessionSegment): string {
    return fallbackMarkLabel(segment, showDate.value);
  }

  /**
   * Activates one recorded-position mark: select its recording and play from that moment.
   *
   * One event, not `select` followed by `play-from`: the page handles both by loading the
   * recording, so emitting both issued two identical points requests, and the first reply could
   * supersede the second and cancel the crossing it was meant to start.
   */
  function playFromMark(segment: SessionSegment, at: number): void {
    emit('play-from', { id: segment.id, at });
  }

  /** Plays from the recorded position nearest a click between marks (keeps click-anywhere). */
  function playFromPointer(segment: SessionSegment, event: MouseEvent): void {
    const element = event.currentTarget as HTMLElement;
    const box = element.getBoundingClientRect();
    const ratio = box.width > 0 ? (event.clientX - box.left) / box.width : 0;
    emit('play-from', { id: segment.id, at: nearestPositionAt(segment, ratio) });
  }
</script>

<template>
  <div class="nh-coverage">
    <div class="nh-coverage-head">
      <span class="nh-coverage-title">
        {{ headline }}
      </span>
      <span class="nh-muted">
        {{ countLabel(totalSamples, 'recorded position') }}<template v-if="shownMarkCount < totalSamples"> · {{ shownMarkCount }} marks shown</template>
      </span>
    </div>
    <div ref="track" class="nh-coverage-track" role="group" aria-label="Recordings in the selected range">
      <!--
        One focusable control per recorded movement. Each mark names its time and recording and
        plays from exactly that moment; the recording's group label states the whole range. Clicking
        the strip between marks still plays from the nearest recorded position.
      -->
      <div
        v-for="entry in rendered"
        :key="entry.segment.id"
        class="nh-coverage-segment"
        :class="{
          'is-selected': entry.segment.id === selectedId,
          'is-pending': entry.segment.id === pendingId && entry.segment.id !== selectedId,
        }"
        role="group"
        :aria-label="label(entry.segment)"
        :aria-current="entry.segment.id === selectedId ? 'true' : undefined"
        :data-session-id="entry.segment.id"
        :data-selected="entry.segment.id === selectedId ? 'true' : 'false'"
        :aria-busy="entry.segment.id === pendingId || undefined"
        :title="label(entry.segment)"
        :style="{
          left: `${entry.leftPct}%`,
          width: `${entry.widthPct}%`,
        }"
        @click="playFromPointer(entry.segment, $event)">
        <button
          v-for="mark in entry.marks"
          :key="`${entry.segment.id}-${mark.at}`"
          type="button"
          class="nh-coverage-mark"
          :data-at="mark.at"
          :data-key="markKey(entry.segment.id, mark.at)"
          :tabindex="tabStopKey === markKey(entry.segment.id, mark.at) ? 0 : -1"
          :style="{ left: markLeft(mark.pct) }"
          :aria-label="markLabel(entry.segment, mark.at)"
          :title="markLabel(entry.segment, mark.at)"
          @focus="activeMarkKey = markKey(entry.segment.id, mark.at)"
          @keydown="onMarkKeydown($event, markKey(entry.segment.id, mark.at))"
          @click.stop="playFromMark(entry.segment, mark.at)" />
        <!--
          No marks (empty recording or a tight global budget): one selection control, but it must
          not paint a tick — that would present an invented recorded position as measured.
        -->
        <button
          v-if="entry.fallback"
          type="button"
          class="nh-coverage-mark is-fallback"
          style="left: 50%"
          :data-key="markKey(entry.segment.id, null)"
          :tabindex="tabStopKey === markKey(entry.segment.id, null) ? 0 : -1"
          :aria-label="fallbackLabel(entry.segment)"
          @focus="activeMarkKey = markKey(entry.segment.id, null)"
          @keydown="onMarkKeydown($event, markKey(entry.segment.id, null))"
          @click.stop="playFromMark(entry.segment, entry.segment.from)" />
      </div>
    </div>
    <div class="nh-coverage-ticks" aria-hidden="true">
      <span v-for="(tick, index) in ticks" :key="index">{{ edgeLabel(tick) }}</span>
    </div>
    <p v-if="!segments.length && loading" class="nh-muted nh-coverage-empty" role="status">
      {{ RECORDINGS_LOADING_COPY }}
    </p>
    <p v-else-if="!segments.length" class="nh-muted nh-coverage-empty">
      {{ RANGE_EMPTY_COPY }}
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
    font-size: 0.78rem;
  }
  .nh-coverage-title {
    font-weight: 600;
    font-variant-numeric: tabular-nums;
  }
  /* A short strip: it is an overview of when things happened, not a second scrubber. */
  .nh-coverage-track {
    position: relative;
    height: 1.5rem;
    border-radius: var(--nh-radius-m);
    background: var(--nh-surface-2);
    box-shadow: inset 0 0 0 1px var(--nh-border);
    overflow: hidden;
  }
  /*
   * Unselected recordings are quiet outlines, never fills: the strip is a few pixels tall, and a
   * filled block would compete with the selected recording for the eye.
   */
  .nh-coverage-segment {
    position: absolute;
    top: 0.2rem;
    bottom: 0.2rem;
    /* A two-second recording inside a three-hour envelope is one pixel wide; hiding it would
       misreport that the time was empty. The floor is on the drawn box only, never on its time. */
    min-width: 3px;
    border-radius: 2px;
    background: transparent;
    box-shadow: inset 0 0 0 1px var(--nh-border-strong);
    overflow: visible;
  }
  /* Selected: the one solid block on the strip, plus a leading edge bar. */
  .nh-coverage-segment.is-selected {
    top: 0;
    bottom: 0;
    background: var(--nh-accent);
  }
  .nh-coverage-segment.is-pending {
    box-shadow: inset 0 0 0 2px var(--nh-border-strong);
  }
  /* Edge bar sits inside the wrapper so the selected block still reads as one shape. */
  .nh-coverage-segment.is-selected::after {
    content: '';
    position: absolute;
    inset-block: 0;
    inset-inline-start: 0;
    width: 3px;
    background: var(--nh-accent-contrast);
    border-radius: 2px;
  }
  /*
   * Every recorded position is its own focusable control. The visible tick stays thin, but the
   * button carries a padded hit area. Colours come from the theme tokens: muted on a quiet
   * recording, the contrast token on the accent-filled selected recording.
   */
  .nh-coverage-mark {
    position: absolute;
    top: 50%;
    height: 1.5rem; /* 24 px target height, centred on the 1.5 rem track */
    width: 1.1rem;
    padding: 0;
    border: 0;
    background: none;
    transform: translate(-50%, -50%);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    /*
     * Dense marks overlap at the 28 px floor, so a pointer click would hit whichever button
     * happened to be last in the DOM rather than the tick under the cursor. Marks stay focusable
     * and Enter/Space-activatable; pointer seeking falls through to the segment's
     * nearest-recorded-position handler, which lands on the tick under the pointer.
     */
    pointer-events: none;
  }
  .nh-coverage-segment {
    cursor: pointer;
  }
  .nh-coverage-mark::before {
    content: '';
    width: 2px;
    height: 62%;
    border-radius: 1px;
    background: var(--nh-text-muted);
  }
  /* No samples: a selection target, not a painted position. */
  .nh-coverage-mark.is-fallback::before {
    content: none;
  }
  .nh-coverage-mark:hover::before {
    background: var(--nh-text);
  }
  .nh-coverage-mark:focus-visible {
    outline: 2px solid var(--nh-focus);
    outline-offset: -1px;
    border-radius: 2px;
  }
  .nh-coverage-mark:focus-visible::before {
    background: var(--nh-accent);
    width: 3px;
  }
  .nh-coverage-segment.is-selected .nh-coverage-mark::before {
    background: var(--nh-accent-contrast);
  }
  /*
   * --nh-focus resolves to --nh-accent, the selected segment's fill, so the ring would be
   * invisible exactly where the focus is. Switch to the contrast token on the accent fill.
   */
  .nh-coverage-segment.is-selected .nh-coverage-mark:focus-visible {
    outline-color: var(--nh-accent-contrast);
  }
  .nh-coverage-segment.is-selected .nh-coverage-mark:hover::before,
  .nh-coverage-segment.is-selected .nh-coverage-mark:focus-visible::before {
    background: var(--nh-accent-contrast);
    width: 3px;
  }
  .nh-coverage-ticks {
    display: flex;
    justify-content: space-between;
    gap: 0.5rem;
    color: var(--nh-text-muted);
    font-size: 0.78rem;
    font-variant-numeric: tabular-nums;
  }
  .nh-coverage-empty {
    margin: 0;
    font-size: 0.85rem;
  }
  @media (max-width: 899px) {
    /* Marks are pointer-transparent and the segment owns touch targeting, so the box stays narrow. */
    .nh-coverage-track {
      height: 1.75rem;
    }
  }
</style>
