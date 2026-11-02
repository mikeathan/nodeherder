import { nextTick, onBeforeUnmount, ref, watch, type WatchSource } from 'vue';

/** Fit natural-height Home blocks into shared grid rows. One page owns the observer, queued
 * measurements and animation frame; replacement/unmount releases every observed DOM target. */
export function useHomeGrid(layoutChanges: WatchSource) {
  const root = ref<HTMLElement | null>(null);
  const measurements = new Map<HTMLElement, number>();
  let pendingFrame = 0;
  let disposed = false;
  let measurementRevision = 0;

  function cancelPendingFrame() {
    if (pendingFrame) cancelAnimationFrame(pendingFrame);
    pendingFrame = 0;
    measurements.clear();
  }

  function applyMeasuredSpans() {
    pendingFrame = 0;
    if (disposed) return;
    const grid = root.value?.querySelector<HTMLElement>('.nh-areas');
    if (!grid) {
      measurements.clear();
      return;
    }
    const style = getComputedStyle(grid);
    const rowHeight = parseFloat(style.gridAutoRows);
    const rowGap = parseFloat(style.rowGap);
    if (!Number.isFinite(rowHeight) || !Number.isFinite(rowGap) || rowHeight + rowGap <= 0) {
      measurements.clear();
      return;
    }
    const firstRowTop = grid.getBoundingClientRect().top;
    const positions = Array.from(grid.children, (element) => {
      const block = element as HTMLElement;
      return { block, firstRow: Math.abs(block.getBoundingClientRect().top - firstRowTop) < 1 };
    });
    let changed = false;
    for (const [block, height] of measurements) {
      if (!block.isConnected || height <= 0) continue;
      const rows = Math.ceil((height + rowGap) / (rowHeight + rowGap));
      const span = `span ${rows}`;
      if (block.style.gridRowEnd !== span) {
        block.style.gridRowEnd = span;
        changed = true;
      }
    }
    measurements.clear();
    for (const { block, firstRow } of positions) {
      if (block.classList.contains('is-area')) continue;
      if (block.classList.contains('is-first-row') !== firstRow) {
        block.classList.toggle('is-first-row', firstRow);
        changed = true;
      }
    }
    // Row spans can move later blocks; reassess their positions after the grid reflows.
    if (changed && !pendingFrame) pendingFrame = requestAnimationFrame(applyMeasuredSpans);
  }

  const observer = new ResizeObserver((entries) => {
    if (disposed) return;
    for (const entry of entries) {
      if (entry.target.classList.contains('nh-areas')) continue;
      const block = entry.target.parentElement;
      if (!block) continue;
      // The border box includes the block's spacing; contentRect excludes its padding.
      const height = entry.borderBoxSize?.[0]?.blockSize ?? entry.target.getBoundingClientRect().height;
      measurements.set(block, height);
    }
    if (!pendingFrame) pendingFrame = requestAnimationFrame(applyMeasuredSpans);
  });

  async function observeRenderedBlocks() {
    const revision = ++measurementRevision;
    await nextTick();
    if (disposed || revision !== measurementRevision) return;
    observer.disconnect();
    cancelPendingFrame();
    const grid = root.value?.querySelector<HTMLElement>('.nh-areas');
    if (grid) observer.observe(grid, { box: 'border-box' });
    const contents = root.value?.querySelectorAll<HTMLElement>('.nh-home-block-content');
    // Batch natural-height reads before writing spans, including the first paint after editing.
    for (const content of Array.from(contents ?? [])) {
      observer.observe(content, { box: 'border-box' });
      const block = content.parentElement;
      if (block) measurements.set(block, content.getBoundingClientRect().height);
    }
    applyMeasuredSpans();
  }

  const stopWatching = watch([root, layoutChanges], observeRenderedBlocks, { immediate: true, flush: 'post' });
  onBeforeUnmount(() => {
    disposed = true;
    measurementRevision++;
    stopWatching();
    observer.disconnect();
    cancelPendingFrame();
  });
  return root;
}
