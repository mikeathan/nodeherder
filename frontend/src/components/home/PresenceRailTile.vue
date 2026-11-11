<script setup lang="ts">
  import { ref } from 'vue';
  import type { Device } from '@/types/device';
  import { RouteName } from '@/types/router';
  import { statusOf } from '@/composables/useHub';
  import { useInViewport } from '@/composables/useInViewport';
  import { usePresenceTile } from '@/composables/spatial/usePresenceTile';
  import { PRESENCE_COPY } from '@/domain/spatial/presence';
  import { SpatialViewIds } from '@/domain/spatial/view';
  import SpatialCanvas from '@/components/device/spatial/SpatialCanvas.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';

  const props = withDefaults(defineProps<{ device: Device; showName?: boolean }>(), { showName: true });
  const root = ref<HTMLElement | null>(null);
  const visible = useInViewport(root);
  const { preview, values, stateLabel, targetLabel, zoneLabel, referenceLabel } = usePresenceTile(
    () => props.device,
    () => visible.value && !document.hidden
  );
</script>

<template>
  <section ref="root" class="nh-sensor-panel" :aria-label="PRESENCE_COPY.tile.ariaLabel(device.friendly_name)">
    <div class="nh-sensor-preview" :class="{ 'has-frame': preview.hasFrameData.value }">
      <SpatialCanvas v-if="preview.hasFrameData.value" :workspace="preview.workspace.value" :tracks="preview.tracks.value"
        :occupancy="new Map()" :view-box="preview.viewBox.value" :scale="preview.scale.value" :orientation="0"
        :online="preview.online.value" :mode="SpatialViewIds.Live" selected-id="" :trails="[]" :cells="[]"
        :show-grid="false" :show-room-context="true" :show-target-labels="false" :target-scale="2.5"
        :label="PRESENCE_COPY.tile.previewLabel(device.friendly_name)" />
      <UiIcon v-else name="presence" />
    </div>
    <div class="nh-sensor-details">
      <header v-if="showName" class="nh-sensor-head">
        <strong class="nh-sensor-name">{{ device.friendly_name }}</strong>
        <span class="nh-muted nh-sensor-status">{{ statusOf(device) }}</span>
      </header>
      <p class="nh-sensor-state">
        <span class="nh-sensor-state-label" :class="{ 'is-occupied': values.occupied }">{{ stateLabel }}</span>
        <span v-if="targetLabel" class="nh-muted"> · {{ targetLabel }}</span>
        <span v-if="zoneLabel" class="nh-muted"> · {{ zoneLabel }}</span>
      </p>
      <div class="nh-sensor-foot">
        <span class="nh-muted nh-sensor-reference">{{ referenceLabel }}</span>
        <span class="nh-sensor-actions">
          <RouterLink
            v-for="action in PRESENCE_COPY.tile.actions"
            :key="action.label"
            :to="{ name: RouteName.DevicePage, params: { id: device.id }, query: { tab: action.tab } }">
            {{ action.label }}
          </RouterLink>
        </span>
      </div>
    </div>
  </section>
</template>

<style scoped>
  .nh-sensor-panel {
    display: grid;
    grid-template-columns: 3.75rem minmax(0, 1fr);
    align-items: center;
    gap: 0.6rem;
    min-width: 0;
    padding: 0.55rem 0;
  }
  .nh-sensor-panel + .nh-sensor-panel {
    border-top: 1px solid var(--nh-border);
  }
  .nh-sensor-preview {
    display: grid;
    place-items: center;
    width: 3.75rem;
    aspect-ratio: 5 / 3;
    flex: none;
    overflow: hidden;
    border-radius: var(--nh-radius-s);
    background: var(--nh-surface-2);
    color: var(--nh-text-muted);
  }
  .nh-sensor-preview .nh-ic {
    width: 1.25rem;
    height: 1.25rem;
  }
  .nh-sensor-preview :deep(svg) {
    display: block;
    width: 100%;
    height: 100%;
    aspect-ratio: auto;
    max-height: none;
  }
  .nh-sensor-details {
    display: grid;
    gap: 0.1rem;
    min-width: 0;
  }
  .nh-sensor-head {
    display: flex;
    align-items: baseline;
    gap: 0.4rem;
    min-width: 0;
  }
  .nh-sensor-name {
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--nh-text);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .nh-sensor-status,
  .nh-sensor-reference {
    font-size: 0.7rem;
  }
  .nh-sensor-state {
    margin: 0;
    font-size: 0.82rem;
    line-height: 1.3;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .nh-sensor-state-label {
    font-weight: 500;
    color: var(--nh-text-muted);
  }
  .nh-sensor-state-label.is-occupied {
    color: var(--nh-accent);
  }
  .nh-sensor-foot {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: 0.5rem;
    flex-wrap: wrap;
    min-width: 0;
  }
  .nh-sensor-reference {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .nh-sensor-actions {
    display: inline-flex;
    gap: 0.75rem;
    font-size: 0.7rem;
    white-space: nowrap;
  }
</style>
