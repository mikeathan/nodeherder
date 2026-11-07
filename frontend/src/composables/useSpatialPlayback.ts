import { onScopeDispose, ref } from 'vue';

/** Playback bounds of the confirmed session, in epoch milliseconds. */
export type PlaybackBounds = { from: number; to: number };

const TICK_MS = 200;

/**
 * Owns the Positions playback clock: the current time, play/pause, speed and the fixed
 * 200 ms tick. Bounds are read on every tick so a session switch cannot leave a stale
 * interval running against the previous recording. The interval is cleared when the
 * owning scope is disposed.
 */
export function useSpatialPlayback(bounds: () => PlaybackBounds | null) {
  const time = ref(0);
  const playing = ref(false);
  const speed = ref(1);
  let timer: ReturnType<typeof setInterval> | null = null;

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
    time.value = Math.min(range.to, time.value + TICK_MS * speed.value);
    if (time.value >= range.to) stop();
  }

  function play() {
    const range = bounds();
    if (playing.value || !range) return;
    if (time.value >= range.to) time.value = range.from;
    playing.value = true;
    timer = setInterval(advance, TICK_MS);
  }

  function seek(value: number) {
    stop();
    time.value = value;
  }

  function toggle() {
    if (playing.value) stop();
    else play();
  }

  onScopeDispose(stop);

  return { time, playing, speed, play, pause: stop, toggle, seek };
}
