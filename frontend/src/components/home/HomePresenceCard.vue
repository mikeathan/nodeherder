<script setup lang="ts">
  import { computed, ref } from 'vue';
  import type { HomePresenceCard } from '@/types/home';
  import { RouteName } from '@/types/router';
  import { findDevice, statusOf } from '@/composables/useHub';
  import { useDerivedPresence } from '@/composables/spatial/useDerivedPresence';
  import { useInViewport } from '@/composables/useInViewport';
  import { useSpatialPreview } from '@/composables/spatial/useSpatialPreview';
  import { formatTime } from '@/utils/date.utils';
  import { derivedPresenceDisplay } from '@/domain/spatial/derived';
  import {
    PRESENCE_COPY,
    presenceEmptyLabel,
    presenceStateLabel,
    presenceTargetLabel,
    presenceZoneChips,
    presenceZoneLabel,
  } from '@/domain/spatial/presence';
  import { SpatialViewIds } from '@/domain/spatial/view';
  import UiButton from '@/components/ui/UiButton.vue';
  import SpatialCanvas from '@/components/device/spatial/SpatialCanvas.vue';

  const props = defineProps<{ card: HomePresenceCard; editing?: boolean; busy?: boolean }>();
  defineEmits<{ remove: [] }>();
  const device = computed(() => findDevice(props.card.device_id));
  const derived = useDerivedPresence(() => device.value);
  const zones = computed(() => presenceZoneChips(derived.derived.value));
  const status = computed(() => (device.value ? statusOf(device.value) : 'missing'));
  const values = computed(() => derived.values.value);
  const display = computed(() => derivedPresenceDisplay(values.value));
  const stateLabel = computed(() => presenceStateLabel(display.value, 'card'));
  const targetLabel = computed(() => presenceTargetLabel(values.value.targetCount));
  const zoneLabel = computed(() => presenceZoneLabel(values.value.lastZone));
  const emptyLabel = computed(() =>
    display.value === 'clear' ? presenceEmptyLabel(values.value.emptySeconds) : null
  );
  const livePreview = ref(false);
  const cardRoot = ref<HTMLElement | null>(null);
  const visible = useInViewport(cardRoot);
  const preview = useSpatialPreview(
    () => props.card.device_id,
    () => livePreview.value && visible.value && !document.hidden
  );
  const updatedLabel = computed(() => (preview.updatedAt.value ? formatTime(preview.updatedAt.value) : ''));
</script>

<template>
  <section ref="cardRoot" class="nh-presence-card" :aria-label="PRESENCE_COPY.card.ariaLabel(device?.friendly_name ?? card.device_id)">
    <header>
      <div>
        <strong>{{ device?.friendly_name ?? PRESENCE_COPY.card.unknownDevice }}</strong>
        <span class="nh-muted"> · {{ status }}</span>
      </div>
      <label v-if="!editing" class="nh-presence-live">
        <input v-model="livePreview" type="checkbox" /> {{ PRESENCE_COPY.card.live }}
      </label>
      <UiButton
        v-if="editing"
        size="sm"
        variant="ghost"
        icon="delete"
        :label="PRESENCE_COPY.card.remove"
        :disabled="busy"
        @click="$emit('remove')" />
    </header>
    <p v-if="!device" class="nh-muted">{{ PRESENCE_COPY.card.missing }}</p>
    <template v-else>
      <p v-if="!derived.derived.value" class="nh-muted">{{ PRESENCE_COPY.card.notDerived }}</p>
      <template v-else>
        <p class="nh-presence-state">
          <strong>{{ stateLabel }}</strong>
          <span v-if="targetLabel"> · {{ targetLabel }}</span>
          <span v-if="zoneLabel"> · {{ zoneLabel }}</span>
          <span v-if="emptyLabel"> · {{ emptyLabel }}</span>
        </p>
        <p v-if="display === 'not-derived'" class="nh-muted">{{ PRESENCE_COPY.card.notDerivedNote }}</p>
        <ul v-if="zones.length" class="nh-presence-zones" :aria-label="PRESENCE_COPY.card.zoneOccupancy">
          <li v-for="zone in zones" :key="zone.id" :class="{ 'is-occupied': zone.occupied }">
            {{ zone.name }}: {{ zone.label }}
          </li>
        </ul>
        <div v-if="livePreview" class="nh-presence-preview">
          <SpatialCanvas v-if="preview.hasFrame.value" :workspace="preview.workspace.value" :tracks="preview.tracks.value"
            :occupancy="new Map()" :scale="4000" :orientation="0" :online="preview.online.value"
            :mode="SpatialViewIds.Live" selected-id="" :trails="[]" :cells="[]" :show-grid="false"
            :show-room-context="false" :show-target-labels="false" :label="PRESENCE_COPY.card.previewLabel" />
          <p v-else class="nh-muted">{{ PRESENCE_COPY.card.waiting }}</p>
          <span v-if="updatedLabel" class="nh-muted">{{ PRESENCE_COPY.card.livePreview }} · {{ updatedLabel }}</span>
        </div>
        <RouterLink :to="{ name: RouteName.DevicePage, params: { id: card.device_id } }">
          {{ PRESENCE_COPY.card.openDevice }}
        </RouterLink>
      </template>
    </template>
  </section>
</template>

<style scoped>
  .nh-presence-card {
    display: grid;
    gap: 0.4rem;
    padding: 0.75rem;
    border: 1px solid var(--nh-border);
    border-radius: var(--nh-radius);
    background: var(--nh-surface);
    min-width: 0;
  }
  header {
    display: flex;
    justify-content: space-between;
    gap: 0.5rem;
    align-items: start;
  }
  .nh-presence-live {
    display: inline-flex;
    gap: 0.3rem;
    align-items: center;
    font-size: 0.85em;
    white-space: nowrap;
  }
  .nh-presence-preview {
    display: grid;
    gap: 0.25rem;
    min-width: 0;
  }
  .nh-presence-preview :deep(svg) {
    width: 100%;
    height: auto;
    max-height: 11rem;
  }
  .nh-presence-state {
    margin: 0;
  }
  .nh-presence-zones {
    display: flex;
    flex-wrap: wrap;
    gap: 0.35rem;
    list-style: none;
    padding: 0;
    margin: 0;
  }
  .nh-presence-zones li {
    padding: 0.15rem 0.45rem;
    border: 1px solid var(--nh-border);
    border-radius: 999px;
    font-size: 0.85em;
  }
  .nh-presence-zones li.is-occupied {
    border-color: var(--nh-accent);
  }
</style>
