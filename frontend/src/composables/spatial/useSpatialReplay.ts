import { computed, onBeforeUnmount, ref, watch } from 'vue';
import type { ReplayFrame } from '@/domain/spatial/replay';
import { REPLAY_TICK_MS, frameAt, replayTracks } from '@/domain/spatial/replay';

/** Owns replay seek/playback state over a recorded session; it never touches the store. */
export function useSpatialReplay(frames: () => readonly ReplayFrame[]) {
  const time = ref(0);
  const playing = ref(false);
  const speed = ref(1);
  const bounds = computed(() => ({ first: frames()[0]?.at ?? 0, last: frames().at(-1)?.at ?? 0 }));
  const frame = computed(() => frameAt(frames(), time.value || bounds.value.last));
  const tracks = computed(() => replayTracks(frame.value));
  let timer: ReturnType<typeof setInterval> | null = null;

  function stopTimer() {
    if (timer) clearInterval(timer);
    timer = null;
  }
  function pause() {
    playing.value = false;
    stopTimer();
  }
  function seek(at: number) {
    const { first, last } = bounds.value;
    time.value = Math.min(Math.max(at, first), last);
  }
  function step() {
    const { last } = bounds.value;
    time.value += REPLAY_TICK_MS * speed.value;
    if (time.value >= last) {
      time.value = last;
      pause();
    }
  }
  function play() {
    if (playing.value || !frames().length) return;
    if (time.value >= bounds.value.last) time.value = bounds.value.first;
    playing.value = true;
    stopTimer();
    timer = setInterval(step, REPLAY_TICK_MS);
  }
  watch(
    frames,
    () => {
      if (!playing.value) time.value = bounds.value.last;
    },
    { immediate: true }
  );
  onBeforeUnmount(pause);

  return { time, playing, speed, bounds, frame, tracks, seek, play, pause };
}
