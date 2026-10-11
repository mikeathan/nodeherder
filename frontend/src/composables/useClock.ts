/* Shared clocks: one timer per interval for the whole app, however many components read it. */
import { onScopeDispose, ref, Ref } from 'vue';

const clocks = new Map<number, { now: Ref<number>; users: number; id: ReturnType<typeof setInterval> | null }>();

/** Milliseconds since epoch, refreshed every `intervalMs` while any component uses it. */
export function useNow(intervalMs = 30_000): Ref<number> {
  let clock = clocks.get(intervalMs);
  if (!clock) {
    clock = { now: ref(Date.now()), users: 0, id: null };
    clocks.set(intervalMs, clock);
  }
  const c = clock;
  if (c.users++ === 0) {
    c.now.value = Date.now();
    c.id = setInterval(() => (c.now.value = Date.now()), intervalMs);
  }
  onScopeDispose(() => {
    if (--c.users === 0 && c.id) {
      clearInterval(c.id);
      c.id = null;
    }
  });
  return c.now;
}
