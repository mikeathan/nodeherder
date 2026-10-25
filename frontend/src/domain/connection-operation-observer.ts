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

/** Owns the timer and request cancellation for the latest local progress observation. */
export function createOperationObserver(options: OperationObserverOptions) {
  let current: AbortController | undefined;
  let timer: ReturnType<typeof setTimeout> | undefined;
  const now = options.now ?? Date.now;

  function stop() {
    current?.abort();
    current = undefined;
    if (timer !== undefined) clearTimeout(timer);
    timer = undefined;
  }

  function start(id: string) {
    stop();
    const controller = new AbortController();
    current = controller;
    const deadline = now() + 30_000;
    const active = () => current === controller && !controller.signal.aborted;

    async function tick() {
      try {
        const operation = await options.read(id, controller.signal);
        if (!active()) return;
        await options.refresh(controller.signal);
        if (active() && ['queued', 'running'].includes(operation.status) && now() < deadline) {
          timer = setTimeout(() => void tick(), 700);
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
