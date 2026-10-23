import type { Plugin, Chart } from 'chart.js';
import { clampHistoryRange, panHistoryRange, zoomHistoryRange, type HistoryRange } from '@/domain/history/range';

export type HistoryInteraction = {
  range: () => HistoryRange;
  bounds: () => HistoryRange;
  change: (range: HistoryRange) => void;
  readout: (timestamp: number | null) => void;
};
export function keyboardHistoryRange(key: string, range: HistoryRange, bounds: HistoryRange): HistoryRange | null {
  if (key === '+' || key === '=') return zoomHistoryRange(range, bounds, 0.5);
  if (key === '-') return zoomHistoryRange(range, bounds, 2);
  if (key === '0') return { ...bounds };
  if (key === 'ArrowLeft' || key === 'ArrowRight') return panHistoryRange(range, bounds, key === 'ArrowLeft' ? -1 : 1);
  return null;
}
export function pinchHistoryRange(range: HistoryRange, bounds: HistoryRange, ratio: number, anchor: number) {
  return zoomHistoryRange(range, bounds, Math.max(0.01, ratio), anchor);
}

/** One chart owns its listeners, captured pointers and one coalesced animation frame. */
export function historyInteractionPlugin(control: HistoryInteraction): Plugin<'line' | 'bar'> {
  let cleanup = () => {};
  let marker: number | null = null;
  return {
    id: 'history-interactions',
    afterInit(chart) {
      cleanup = bindInteractions(chart, { ...control, readout(at) { marker = at; control.readout(at); chart.draw(); } });
    },
    beforeUpdate() { marker = null; },
    afterDraw(chart) {
      if (marker === null) return;
      const x = chart.scales.x.getPixelForValue(marker), area = chart.chartArea;
      if (x < area.left || x > area.right) return;
      const ctx = chart.ctx;
      ctx.save();
      ctx.strokeStyle = getComputedStyle(chart.canvas).color;
      ctx.setLineDash([3, 3]);
      ctx.beginPath(); ctx.moveTo(x, area.top); ctx.lineTo(x, area.bottom); ctx.stroke();
      ctx.restore();
    },
    beforeDestroy() { cleanup(); },
  };
}

function bindInteractions(chart: Chart<'line' | 'bar'>, control: HistoryInteraction) {
  const canvas = chart.canvas;
  const host = canvas.closest<HTMLElement>('[data-history-chart]');
  const touches = new Map<number, { x: number; y: number }>();
  let drag: { id: number; x: number } | null = null;
  let pinch: { distance: number; anchor: number; range: HistoryRange } | null = null;
  let multiTouch = false;
  let frame = 0;
  let pendingRange: HistoryRange | null = null;
  let pendingTime: number | null = null;
  let lastTap = 0;
  let moved = false;
  const fraction = (clientX: number) => {
    const area = chart.chartArea;
    return Math.max(0, Math.min(1, (clientX - canvas.getBoundingClientRect().left - area.left) / (area.right - area.left)));
  };
  const timestamp = (clientX: number, range = control.range()) => range.from + fraction(clientX) * (range.to - range.from);
  function flush() {
    frame = 0;
    if (pendingRange) { pendingTime = null; control.change(clampHistoryRange(pendingRange, control.bounds())); }
    if (pendingTime !== null) control.readout(pendingTime);
    pendingRange = null;
    pendingTime = null;
  }
  function schedule(range: HistoryRange | null, time: number | null = null) {
    if (range) pendingRange = range;
    if (time !== null) pendingTime = time;
    if (!frame) frame = requestAnimationFrame(flush);
  }
  function down(event: PointerEvent) {
    if (event.button !== 0) return;
    canvas.setPointerCapture(event.pointerId);
    moved = false;
    if (event.pointerType !== 'touch') { drag = { id: event.pointerId, x: event.clientX }; return; }
    touches.set(event.pointerId, { x: event.clientX, y: event.clientY });
    if (touches.size === 2) {
      multiTouch = true;
      const [a, b] = [...touches.values()];
      pinch = { distance: Math.max(1, Math.hypot(b.x - a.x, b.y - a.y)), anchor: fraction((a.x + b.x) / 2), range: { ...control.range() } };
    } else if (!multiTouch) schedule(null, timestamp(event.clientX));
  }
  function move(event: PointerEvent) {
    if (event.pointerType !== 'touch') { if (!drag) schedule(null, timestamp(event.clientX)); return; }
    if (!touches.has(event.pointerId)) return;
    moved = true;
    touches.set(event.pointerId, { x: event.clientX, y: event.clientY });
    if (touches.size === 2 && pinch) {
      const [a, b] = [...touches.values()];
      schedule(pinchHistoryRange(pinch.range, control.bounds(), pinch.distance / Math.max(1, Math.hypot(b.x - a.x, b.y - a.y)), pinch.anchor));
    } else if (!multiTouch) schedule(null, timestamp(event.clientX));
  }
  function up(event: PointerEvent) {
    if (drag?.id === event.pointerId) {
      if (Math.abs(event.clientX - drag.x) > 8) {
        const a = timestamp(drag.x), b = timestamp(event.clientX);
        schedule({ from: Math.min(a, b), to: Math.max(a, b) });
      } else schedule(null, timestamp(event.clientX));
      drag = null;
    }
    if (touches.has(event.pointerId)) {
      if (!multiTouch && !moved) {
        if (lastTap > 0 && event.timeStamp - lastTap < 300) schedule(control.bounds());
        lastTap = event.timeStamp;
      }
      touches.delete(event.pointerId);
      if (!touches.size) { pinch = null; multiTouch = false; }
    }
    if (canvas.hasPointerCapture(event.pointerId)) canvas.releasePointerCapture(event.pointerId);
  }
  function cancel(event: PointerEvent) {
    if (drag?.id === event.pointerId) drag = null;
    touches.delete(event.pointerId);
    if (!touches.size) { pinch = null; multiTouch = false; }
    pendingRange = null;
  }
  function wheel(event: WheelEvent) {
    if (!event.deltaY) return;
    event.preventDefault();
    // Preserve every wheel delta even while one RAF/Vue update is still queued.
    schedule(zoomHistoryRange(pendingRange ?? control.range(), control.bounds(), event.deltaY < 0 ? 0.8 : 1.25, fraction(event.clientX)));
  }
  function key(event: KeyboardEvent) {
    if (event.target !== host) return;
    const range = keyboardHistoryRange(event.key, control.range(), control.bounds());
    if (range) { event.preventDefault(); schedule(range); }
  }
  function reset() { schedule(control.bounds()); }
  function leave(event: PointerEvent) {
    if (event.pointerType !== 'touch' && !drag) { pendingTime = null; control.readout(null); }
  }
  canvas.addEventListener('pointerdown', down);
  canvas.addEventListener('pointermove', move);
  canvas.addEventListener('pointerup', up);
  canvas.addEventListener('pointercancel', cancel);
  canvas.addEventListener('pointerleave', leave);
  canvas.addEventListener('wheel', wheel, { passive: false });
  canvas.addEventListener('dblclick', reset);
  host?.addEventListener('keydown', key);
  return () => {
    cancelAnimationFrame(frame);
    if (drag && canvas.hasPointerCapture(drag.id)) canvas.releasePointerCapture(drag.id);
    for (const id of touches.keys()) if (canvas.hasPointerCapture(id)) canvas.releasePointerCapture(id);
    canvas.removeEventListener('pointerdown', down);
    canvas.removeEventListener('pointermove', move);
    canvas.removeEventListener('pointerup', up);
    canvas.removeEventListener('pointercancel', cancel);
    canvas.removeEventListener('pointerleave', leave);
    canvas.removeEventListener('wheel', wheel);
    canvas.removeEventListener('dblclick', reset);
    host?.removeEventListener('keydown', key);
  };
}
