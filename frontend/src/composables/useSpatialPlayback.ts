import { computed, onScopeDispose, ref, watch } from 'vue';

/** Playback bounds of the confirmed session, in epoch milliseconds. */
export type PlaybackBounds = { from: number; to: number };

/**
 * Transport to reach the recording that follows the current one, or null at the end of the
 * filtered set. The composable stays free of loading code: the caller performs the switch.
 */
export type PlaybackCrossing = () => { id: string; from: number } | null;

const TICK_MS = 200;

/**
 * Owns the Positions playback clock: the current time, play/pause, speed and the fixed
 * 200 ms tick. Bounds are read on every tick so a session switch cannot leave a stale
 * interval running against the previous recording.
 *
 * Across recordings the caller supplies `crossing`: at the end of the current bounds the
 * clock is held (playing stays true) and the caller performs the load. `reachedEnd` is set
 * only when nothing follows, so the transport can say playback finished instead of looking
 * paused. `awaitingCrossing` marks the in-flight load; `failCrossing()` stops playback and
 * reports why, so a failed transition is never presented as played.
 */
export function useSpatialPlayback(bounds: () => PlaybackBounds | null, crossing: PlaybackCrossing) {
  const time = ref(0);
  const playing = ref(false);
  const speed = ref(1);
  const reachedEnd = ref(false);
  const awaitingCrossing = ref(false);
  let timer: ReturnType<typeof setInterval> | null = null;
  let crossingAt = 0;

  /*
   * The confirmed recording changes while the clock is paused: the first load, a refresh, a
   * range or device change, and a failed switch all leave the playhead at the previous value.
   * Placing it at the new recording's start is a placement, not a seek: it must not stop a clock
   * the operator started (a pending crossing is handled by resumeCrossing instead).
   */
  const range = computed(bounds);
  watch(range, (next) => {
    if (!next || awaitingCrossing.value) return;
    if (time.value < next.from || time.value > next.to) time.value = next.from;
  });

  function stop() {
    playing.value = false;
    if (timer) clearInterval(timer);
    timer = null;
  }

  function advance() {
    const range = bounds();
    if (!range) {
      stop();
      return;
    }
    if (awaitingCrossing.value) return;
    // The confirmed recording can change under a running clock (a recording switch, or a
    // superseded crossing). Clamp instead of trusting every caller to have seeked first, so the
    // playhead is never rendered outside the bounds on screen.
    if (time.value < range.from || time.value > range.to) {
      time.value = range.from;
    }
    const next = time.value + TICK_MS * speed.value;
    if (next < range.to) {
      time.value = next;
      return;
    }
    time.value = range.to;
    const following = crossing();
    if (!following) {
      reachedEnd.value = true;
      stop();
      return;
    }
    // Hold the confirmed recording on screen until its successor loads. Playback stays
    // "playing" so the caller can continue without a second Play press, but the clock does
    // not move: presenting an unloaded recording's time would fabricate a playhead.
    awaitingCrossing.value = true;
    crossingAt = following.from;
  }

  /**
   * The pending successor loaded at `from`; the playhead moves to it. This applies even when the
   * operator paused mid-crossing: the confirmed recording changed, so leaving the clock at the
   * previous recording's end would render a playhead outside the bounds now on screen. A seek
   * already cancelled the crossing, so a late settlement does not move the playhead again.
   */
  function resumeCrossing() {
    if (!awaitingCrossing.value) return;
    awaitingCrossing.value = false;
    time.value = crossingAt;
  }

  /**
   * The pending successor failed: stop and keep the confirmed recording. The caller reports the
   * reason with the rest of its request errors, so there is one place that announces it.
   */
  function failCrossing() {
    awaitingCrossing.value = false;
    stop();
  }

  /**
   * Drops a pending crossing without moving the playhead or reporting a failure. Used when the
   * request is superseded by a newer selection: the caller is about to confirm a different
   * recording, and leaving the flag set would make every later tick return early, freezing the
   * clock with the transport still showing "playing".
   */
  function cancelCrossing() {
    awaitingCrossing.value = false;
  }

  /**
   * Starts the clock from the current playhead.
   *
   * @param rewindAtEnd Replace a playhead sitting at (or past) the end with the recording's start.
   * Pass false when the caller has just placed the playhead deliberately: the last recorded sample
   * of a recording sits exactly on its `to`, so rewinding there turned "play from this mark" into
   * "play the whole recording from its first sample".
   */
  function play(rewindAtEnd = true) {
    const range = bounds();
    if (playing.value || !range) return;
    reachedEnd.value = false;
    if (rewindAtEnd && time.value >= range.to) time.value = range.from;
    playing.value = true;
    timer = setInterval(advance, TICK_MS);
  }

  function seek(value: number) {
    stop();
    awaitingCrossing.value = false;
    reachedEnd.value = false;
    time.value = value;
  }

  function toggle() {
    if (playing.value) stop();
    else play();
  }

  onScopeDispose(stop);

  return {
    time,
    playing,
    speed,
    reachedEnd,
    awaitingCrossing,
    play,
    pause: stop,
    toggle,
    seek,
    resumeCrossing,
    failCrossing,
    cancelCrossing,
  };
}
