/** A page may observe one admitted operation. This never starts or polls a provider job. */
export interface OperationProgress {
  status: string;
}

export interface OperationObserverOptions {
  read: (id: string, signal: AbortSignal) => Promise<OperationProgress>;
  refresh: (signal: AbortSignal) => Promise<unknown>;
  onError: (error: unknown) => void;
  now?: () => number;
}

const OBSERVATION_TIMEOUT_MS = 5 * 60_000;
const FAST_POLL_MS = 1_000;
const SLOW_POLL_MS = 3_000;
const FAST_WINDOW_MS = 30_000;

/** Owns the timer and request cancellation for the latest local progress observation. */
export function createOperationObserver(options: OperationObserverOptions) {
  let current: AbortController | undefined;
  let timer: ReturnType<typeof setTimeout> | undefined;
  const now = options.now ?? Date.now;

  /** Revoke ownership before cancelling the timer/read so late callbacks cannot refresh state. */
  function stop() {
    current?.abort();
    current = undefined;
    if (timer !== undefined) clearTimeout(timer);
    timer = undefined;
  }

  /** Observe one admitted operation for at most five minutes, superseding the previous observation. */
  function start(id: string) {
    stop();
    const controller = new AbortController();
    current = controller;
    const started = now();
    const deadline = started + OBSERVATION_TIMEOUT_MS;
    const active = () => current === controller && !controller.signal.aborted;

    async function tick() {
      try {
        if (typeof document !== 'undefined' && document.hidden) {
          if (active() && now() < deadline) timer = setTimeout(() => void tick(), SLOW_POLL_MS);
          return;
        }
        const operation = await options.read(id, controller.signal);
        if (!active()) return;
        await options.refresh(controller.signal);
        if (active() && ['queued', 'running'].includes(operation.status) && now() < deadline) {
          const delay = now() - started < FAST_WINDOW_MS ? FAST_POLL_MS : SLOW_POLL_MS;
          timer = setTimeout(() => void tick(), delay);
        }
      } catch (error) {
        // Cancellation or a newer observation revokes ownership of all callbacks.
        if (active()) options.onError(error);
      }
    }
    void tick();
  }

  return { start, stop };
}
