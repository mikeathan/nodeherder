import { computed, onScopeDispose, ref, watch, type Ref } from 'vue';
import type { SpatialSession } from '@/services/spatial-history.service';
import {
  gapAfter,
  gapBefore,
  playbackRecordings,
  recordingNeighbours,
  type PlaybackGap,
} from '@/domain/spatial/playback';
import { useSpatialPlayback } from '@/composables/useSpatialPlayback';

/** How long a crossed gap stays in the one status line before normal status returns. */
const SKIPPED_GAP_NOTICE_MS = 5_000;

/**
 * Couples the Positions clock to the recording list and the recording loader.
 *
 * The component owns presentation only; everything that can wedge playback lives here:
 * which recording is confirmed, where the playhead sits, when to cross into the next recording,
 * and how a superseded or failed load settles the clock. The clock itself is
 * `useSpatialPlayback`.
 *
 * @param input.sessions Filtered recordings, ascending; the transport reads the order from it.
 * @param input.confirmed Reactive id of the recording whose points are rendered.
 * @param input.load Confirms a recording; carries the points query. Resolves true when confirmed.
 * @param input.supersededAt Counts replies for replaced requests; a pending crossing is dropped
 * when it moves, because that load will never confirm.
 */
export function usePositionsPlayback(input: {
  sessions: Ref<SpatialSession[]>;
  confirmed: Ref<string>;
  load: (sessionId: string) => Promise<boolean>;
  supersededAt: Ref<number>;
}) {
  const bounds = computed(() => {
    const session = input.sessions.value.find((entry) => entry.id === input.confirmed.value);
    return session ? { from: session.from, to: session.to } : null;
  });

  /** Filtered recordings in playback order. */
  const recordings = computed(() => playbackRecordings(input.sessions.value));
  /** Position among the filtered recordings and their adjacent ids. */
  const neighbours = computed(() =>
    recordingNeighbours(input.sessions.value, input.confirmed.value)
  );
  const nextRecording = computed(
    () => recordings.value.find((recording) => recording.id === neighbours.value.next) ?? null
  );

  const playback = useSpatialPlayback(
    () => bounds.value,
    () => {
      const following = nextRecording.value;
      return following ? { id: following.id, from: following.from } : null;
    }
  );
  const { time, playing, speed, reachedEnd, awaitingCrossing } = playback;

  /** Gap the last continuous advance skipped, so the transport can say so. */
  const skippedGap = ref<PlaybackGap | null>(null);
  let skippedTimer: ReturnType<typeof setTimeout> | null = null;

  /**
   * Show a skipped gap for a few seconds, then yield the status line back to the playhead.
   *
   * The gap is a transient explanation of the boundary just crossed, not a permanent state: left
   * set, it outranks every later idle, pending-gap and end message for the rest of the recording.
   */
  function showSkippedGap(gap: PlaybackGap | null): void {
    if (skippedTimer) clearTimeout(skippedTimer);
    skippedGap.value = gap;
    skippedTimer = gap
      ? setTimeout(() => {
          skippedGap.value = null;
          skippedTimer = null;
        }, SKIPPED_GAP_NOTICE_MS)
      : null;
  }

  onScopeDispose(() => {
    if (skippedTimer) clearTimeout(skippedTimer);
  });
  /** Gap the next continuous advance will skip, previewed before it happens. */
  const pendingGap = computed(() =>
    gapAfter(recordings.value, input.confirmed.value)
  );

  /**
   * Confirm a recording and place the playhead at its start, paused.
   *
   * The playhead must move with the recording: a load that leaves the clock where the previous
   * recording ended renders nothing, because every sample is in the future.
   *
   * @param sessionId Recording to confirm.
   * @param at Instant to seek to; defaults to the recording's first sample.
   * @returns Whether the recording is now confirmed; a failed load keeps the previous content.
   */
  async function select(sessionId: string, at?: number): Promise<boolean> {
    showSkippedGap(null);
    const loaded = await input.load(sessionId);
    if (!loaded) return false;
    playback.pause();
    playback.seek(at ?? bounds.value?.from ?? 0);
    return true;
  }

  /**
   * Start playback at a recording's first sample, or at `at`.
   *
   * @returns Whether the recording is now confirmed; playback only starts from real content.
   */
  async function play(sessionId: string, at?: number): Promise<boolean> {
    if (!(await select(sessionId, at))) return false;
    // An explicit `at` is a deliberate placement (a movement mark sits on the recording's `to`),
    // so the clock must not rewind to the start to escape "already at the end".
    playback.play(at === undefined);
    return true;
  }

  /**
   * Move to the adjacent recording, keeping the play/pause state.
   *
   * Paging while paused must not start a clock the operator did not ask for, and paging while
   * playing must not require a second Play press.
   */
  async function goTo(id: string | null): Promise<void> {
    if (!id) return;
    const resume = playing.value;
    if (!(await select(id))) return;
    if (resume) playback.play();
  }

  // A replaced load will never confirm, so its crossing must not stay pending: every later tick
  // would return early and freeze the clock while the transport still said "playing".
  watch(input.supersededAt, () => playback.cancelCrossing());

  /**
   * Performs the load the clock requested at the end of a recording. The clock only flips
   * `awaitingCrossing`; the transport work stays here, where the requests live.
   */
  watch(awaitingCrossing, async (waiting) => {
    if (!waiting) return;
    const following = nextRecording.value;
    if (!following) {
      // The successor vanished between the tick and this watcher (the filter changed): there is
      // nothing to cross into, and leaving `awaitingCrossing` set would freeze every later tick.
      playback.cancelCrossing();
      playback.pause();
      return;
    }
    showSkippedGap(gapBefore(recordings.value, following.id));
    const loaded = await input.load(following.id);
    /*
     * The crossing may have been settled while this load was in flight: a newer selection
     * supersedes it (cancelCrossing) and may already be playing. Acting on this stale reply would
     * then `failCrossing()` and stop a clock the operator has since started.
     */
    if (!playback.awaitingCrossing.value) return;
    if (!loaded) {
      // The load failed: stop and keep the confirmed recording. Clearing the flag is what lets a
      // later Play resume instead of returning early forever.
      playback.failCrossing();
      return;
    }
    playback.resumeCrossing();
  });

  return {
    time,
    playing,
    speed,
    reachedEnd,
    awaitingCrossing,
    pendingGap,
    skippedGap,
    neighbours,
    recordings,
    playback,
    select,
    play,
    goTo,
  };
}
