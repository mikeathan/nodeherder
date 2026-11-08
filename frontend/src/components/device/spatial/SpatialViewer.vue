<script setup lang="ts">
  /*
   * Shared spatial viewer shell: a collapsible view-control panel above the canvas plus
   * pointer/touch panning. Zoom is owned by the caller (it also drives fitted views); this
   * component applies it to the base viewBox and exposes the effective box to the canvas slot.
   */
  import { computed, ref, useId, watch } from 'vue';
  import { useMediaQuery } from '@/composables/useMediaQuery';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';

  const ZOOM_LEVELS = [1, 2, 3, 5] as const;

  const props = withDefaults(
    defineProps<{
      /** View box fitted by the caller before zoom and pan. */
      baseViewBox: string;
      zoom: number;
      showGrid: boolean;
      showFov: boolean;
      /** Editing canvases need an explicit hand tool; read-only views pan directly. */
      editMode?: boolean;
      /** Optional zoom centre, e.g. the active recorded sample. */
      focus?: { x: number; y: number } | null;
      /** Changes on session or view switches; clears pan. */
      resetKey?: string | number;
    }>(),
    { editMode: false, focus: null, resetKey: '' }
  );
  const emit = defineEmits<{
    'update:zoom': [value: number];
    'update:showGrid': [value: boolean];
    'update:showFov': [value: boolean];
  }>();
  defineSlots<{
    default(props: { viewBox: string; panMode: boolean }): any;
    extra?(): any;
  }>();

  const compact = useMediaQuery('(max-width: 899px)');
  const open = ref(!compact.value);
  const bodyId = useId();
  const viewport = ref<HTMLElement | null>(null);
  const pan = ref({ x: 0, y: 0 });
  const panTool = ref(false);
  const spaceHeld = ref(false);
  const panning = ref(false);

  const base = computed(() => {
    const [x, y, width, height] = props.baseViewBox.split(' ').map(Number);
    return { x, y, width, height };
  });
  const panActive = computed(() => panTool.value || spaceHeld.value || panning.value);
  const pannable = computed(() => !props.editMode || panActive.value);
  const panMode = computed(() => panTool.value || spaceHeld.value || panning.value);

  const viewBox = computed(() => {
    const { x, y, width, height } = base.value;
    const zoom = Math.max(1, props.zoom);
    const zoomWidth = width / zoom;
    const zoomHeight = height / zoom;
    const focus = zoom > 1 ? props.focus : null;
    const centerX = (focus?.x ?? x + width / 2) + pan.value.x;
    const centerY = (focus?.y ?? y + height / 2) + pan.value.y;
    const round = (value: number) => Number(value.toFixed(1));
    return [
      round(centerX - zoomWidth / 2),
      round(centerY - zoomHeight / 2),
      round(zoomWidth),
      round(zoomHeight),
    ].join(' ');
  });

  const summary = computed(() => {
    const parts = [`${props.zoom}×`, props.showGrid ? 'Grid' : 'No grid'];
    if (props.showFov) parts.push('FOV');
    parts.push(props.editMode ? (panTool.value ? 'Pan' : 'Edit') : 'Drag to pan');
    return parts.join(' · ');
  });

  function resetView() {
    pan.value = { x: 0, y: 0 };
    panTool.value = false;
    spaceHeld.value = false;
    emit('update:zoom', 1);
  }

  watch(
    () => props.resetKey,
    () => {
      pan.value = { x: 0, y: 0 };
      panTool.value = false;
      spaceHeld.value = false;
    }
  );

  // A new zoom level or anchor should start centred on that anchor, not on the old pan.
  watch([() => props.zoom, () => props.focus], () => {
    pan.value = { x: 0, y: 0 };
  });

  let drag: { pointerId: number; x: number; y: number; panX: number; panY: number } | null = null;

  function startPan(event: PointerEvent) {
    const element = viewport.value;
    if (!element) return;
    // Keep Space-to-pan working after a canvas click, including in edit mode.
    element.focus({ preventScroll: true });
    if (!pannable.value || event.button !== 0) return;
    drag = {
      pointerId: event.pointerId,
      x: event.clientX,
      y: event.clientY,
      panX: pan.value.x,
      panY: pan.value.y,
    };
    panning.value = true;
    element.setPointerCapture(event.pointerId);
    event.preventDefault();
  }

  function movePan(event: PointerEvent) {
    if (!drag) return;
    const rect = viewport.value?.getBoundingClientRect();
    if (!rect?.width || !rect.height) return;
    const zoom = Math.max(1, props.zoom);
    // preserveAspectRatio "meet" uses one uniform scale; use the limiting dimension so a
    // square live view letterboxed in a wide box pans at the same rate in both axes.
    const pxPerView = Math.min(
      rect.width / (base.value.width / zoom),
      rect.height / (base.value.height / zoom)
    );
    if (!pxPerView) return;
    const viewPerPixel = 1 / pxPerView;
    pan.value = {
      x: drag.panX - (event.clientX - drag.x) * viewPerPixel,
      y: drag.panY - (event.clientY - drag.y) * viewPerPixel,
    };
  }

  function endPan() {
    if (!drag) return;
    const pointerId = drag.pointerId;
    drag = null;
    panning.value = false;
    try {
      viewport.value?.releasePointerCapture(pointerId);
    } catch {
      // Pointer capture may already be gone after pointercancel.
    }
  }

  function onKeydown(event: KeyboardEvent) {
    if (!props.editMode || event.code !== 'Space' || event.repeat) return;
    event.preventDefault();
    spaceHeld.value = true;
  }

  function onKeyup(event: KeyboardEvent) {
    if (event.code === 'Space') spaceHeld.value = false;
  }
</script>

<template>
  <div class="nh-spatial-viewer">
    <div class="nh-view-controls">
      <button
        type="button"
        class="nh-view-toggle"
        aria-label="View controls"
        :aria-expanded="open"
        :aria-controls="bodyId"
        @click="open = !open">
        <UiIcon name="tune" />
        <span class="nh-view-title">View</span>
        <span class="nh-view-summary">{{ summary }}</span>
        <UiIcon :name="open ? 'chevronUp' : 'chevronDown'" />
      </button>
      <div v-show="open" :id="bodyId" class="nh-view-body">
        <div class="nh-view-group">
          <span class="nh-view-label">Zoom</span>
          <div class="nh-view-row">
            <UiButton
              v-for="level in ZOOM_LEVELS"
              :key="level"
              size="sm"
              :variant="zoom === level ? 'primary' : 'ghost'"
              :pressed="zoom === level"
              @click="emit('update:zoom', level)">
              {{ level }}×
            </UiButton>
            <UiButton size="sm" variant="ghost" @click="resetView">Reset view</UiButton>
          </div>
        </div>
        <div class="nh-view-group">
          <span class="nh-view-label">View</span>
          <div class="nh-view-row">
            <label class="nh-view-check">
              <input
                type="checkbox"
                :checked="showGrid"
                @change="emit('update:showGrid', ($event.target as HTMLInputElement).checked)" />
              Grid
            </label>
            <label class="nh-view-check">
              <input
                type="checkbox"
                :checked="showFov"
                @change="emit('update:showFov', ($event.target as HTMLInputElement).checked)" />
              Field of view
            </label>
            <label v-if="editMode" class="nh-view-check">
              <input v-model="panTool" type="checkbox" />
              Pan tool
            </label>
          </div>
        </div>
        <div v-if="$slots.extra" class="nh-view-group nh-view-group--display">
          <span class="nh-view-label">Display</span>
          <div class="nh-view-row"><slot name="extra" /></div>
        </div>
      </div>
    </div>
    <div
      ref="viewport"
      class="nh-viewport"
      :class="{ 'is-pannable': pannable, 'is-panning': panning }"
      tabindex="0"
      role="group"
      aria-label="Spatial view"
      @pointerdown="startPan"
      @pointermove="movePan"
      @pointerup="endPan"
      @pointercancel="endPan"
      @lostpointercapture="endPan"
      @keydown="onKeydown"
      @keyup="onKeyup"
      @dragstart.prevent>
      <slot :viewBox="viewBox" :panMode="panMode" />
    </div>
  </div>
</template>

<style scoped>
  .nh-spatial-viewer {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    min-width: 0;
  }
  .nh-view-controls {
    border: 1px solid var(--nh-border);
    border-radius: var(--nh-radius-m);
    background: var(--nh-surface);
    overflow: hidden;
  }
  .nh-view-toggle {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    width: 100%;
    min-height: 2.5rem;
    padding: 0.35rem 0.65rem;
    border: 0;
    background: transparent;
    color: inherit;
    text-align: start;
    cursor: pointer;
  }
  .nh-view-toggle:hover {
    background: var(--nh-surface-2);
  }
  .nh-view-toggle:focus-visible {
    outline: 2px solid var(--nh-focus);
    outline-offset: -2px;
  }
  .nh-view-title {
    font-weight: 600;
    font-size: 0.85rem;
  }
  .nh-view-summary {
    margin-left: auto;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: var(--nh-text-muted);
    font-size: 0.78rem;
  }
  .nh-view-body {
    display: flex;
    flex-wrap: wrap;
    gap: 0.7rem 1.25rem;
    padding: 0.55rem 0.65rem 0.7rem;
    border-top: 1px solid var(--nh-border);
  }
  .nh-view-group {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
    flex: 1 1 14rem;
    min-width: 0;
  }
  .nh-view-group--display {
    flex-basis: 100%;
  }
  .nh-view-label {
    color: var(--nh-text-muted);
    font-size: 0.72rem;
    font-weight: 600;
    letter-spacing: 0.04em;
    text-transform: uppercase;
  }
  .nh-view-row {
    display: flex;
    flex-wrap: wrap;
    gap: 0.4rem 0.75rem;
    align-items: center;
  }
  .nh-view-row :deep(.nh-field) {
    margin: 0;
  }
  .nh-view-check {
    display: inline-flex;
    align-items: center;
    gap: 0.3rem;
    min-height: 2rem;
    font-size: 0.85rem;
    white-space: nowrap;
  }
  .nh-viewport {
    min-width: 0;
    border-radius: var(--nh-radius-m);
  }
  .nh-viewport.is-pannable {
    cursor: grab;
    touch-action: none;
  }
  .nh-viewport.is-panning {
    cursor: grabbing;
  }
  .nh-viewport:focus-visible {
    outline: 2px solid var(--nh-focus);
    outline-offset: 2px;
  }
  @media (max-width: 899px) {
    .nh-view-summary {
      display: none;
    }
    .nh-view-group {
      flex-basis: 100%;
    }
    .nh-view-check,
    .nh-view-toggle {
      min-height: 44px;
    }
  }
</style>
