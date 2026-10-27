import { effectScope, nextTick, ref } from 'vue';
import { useHomeGrid } from '@/composables/useHomeGrid';

const mockUnmounted: (() => void)[] = [];
jest.mock('vue', () => ({ ...jest.requireActual('vue'), onBeforeUnmount: (fn: () => void) => mockUnmounted.push(fn) }));

test('one observer releases replaced targets, pending frames and queued measurements on unmount', async () => {
  const disconnect = jest.fn(),
    observe = jest.fn();
  let resize!: ResizeObserverCallback;
  const prior = {
    observer: global.ResizeObserver,
    raf: global.requestAnimationFrame,
    cancel: global.cancelAnimationFrame,
    style: global.getComputedStyle,
  };
  global.ResizeObserver = jest.fn().mockImplementation((callback: ResizeObserverCallback) => {
    resize = callback;
    return { observe, disconnect };
  });
  const frames = new Map<number, FrameRequestCallback>();
  global.requestAnimationFrame = jest.fn((callback) => {
    frames.set(1, callback);
    return 1;
  });
  global.cancelAnimationFrame = jest.fn((id) => {
    frames.delete(id);
  });
  global.getComputedStyle = jest.fn().mockReturnValue({ gridAutoRows: '12px', rowGap: '12px' });
  const block = { style: { gridRowEnd: '' }, isConnected: true };
  const content = { parentElement: block, classList: { contains: () => false }, getBoundingClientRect: () => ({ height: 100 }) };
  const grid = { children: [], getBoundingClientRect: () => ({ top: 0 }) };
  const scope = effectScope();
  try {
    const change = ref(0);
    const root = scope.run(() => useHomeGrid(change))!;
    root.value = { querySelector: () => grid, querySelectorAll: () => [content] } as unknown as HTMLElement;
    await nextTick();
    await nextTick();
    expect(observe).toHaveBeenCalled();
    const initialDisconnects = disconnect.mock.calls.length;
    expect(block.style.gridRowEnd).toBe('span 5');
    // The observed content box excludes padding, which is part of the visible block spacing.
    resize(
      [{ target: content, contentRect: { height: 80 } }] as unknown as ResizeObserverEntry[],
      {} as ResizeObserver
    );
    const frame = frames.get(1)!;
    frames.delete(1);
    frame(0);
    expect(block.style.gridRowEnd).toBe('span 5');
    resize(
      [{ target: content, contentRect: { height: 220 } }] as unknown as ResizeObserverEntry[],
      {} as ResizeObserver
    );
    expect(frames.size).toBe(1);
    change.value++;
    await nextTick();
    await nextTick();
    expect(frames.size).toBe(0);
    expect(disconnect).toHaveBeenCalledTimes(initialDisconnects + 1);
    resize(
      [{ target: content, contentRect: { height: 220 } }] as unknown as ResizeObserverEntry[],
      {} as ResizeObserver
    );
    mockUnmounted.splice(0).forEach((fn) => fn());
    scope.stop();
    expect(frames.size).toBe(0);
    expect(disconnect).toHaveBeenCalledTimes(initialDisconnects + 2);
    expect(block.style.gridRowEnd).toBe('span 5');
    resize([], {} as ResizeObserver);
    expect(frames.size).toBe(0);
  } finally {
    scope.stop();
    global.ResizeObserver = prior.observer;
    global.requestAnimationFrame = prior.raf;
    global.cancelAnimationFrame = prior.cancel;
    global.getComputedStyle = prior.style;
  }
});
