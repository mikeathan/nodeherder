<script setup lang="ts">
  import { computed } from 'vue';
  import type { HomePresenceCard } from '@/types/home';
  import { RouteName } from '@/types/router';
  import { findDevice, statusOf } from '@/composables/useHub';
  import { useDerivedPresence } from '@/composables/spatial/useDerivedPresence';
  import { derivedZoneSignals } from '@/domain/spatial/derived';
  import UiButton from '@/components/ui/UiButton.vue';

  const props = defineProps<{ card: HomePresenceCard; editing?: boolean; busy?: boolean }>();
  defineEmits<{ remove: [] }>();
  const device = computed(() => findDevice(props.card.device_id));
  const derived = useDerivedPresence(() => device.value);
  const zones = computed(() => {
    const signals = derivedZoneSignals(derived.derived.value);
    return [...signals.entries()]
      .map(([zoneId, signal]) => {
        const entity = Object.values(derived.derived.value?.exposes ?? {}).find(
          (candidate) => candidate.attributes?.zoneId === zoneId && candidate.name.endsWith('_occupied')
        );
        return {
          id: zoneId,
          name: typeof entity?.attributes?.zoneName === 'string' ? entity.attributes.zoneName : zoneId,
          occupied: signal.occupied,
        };
      })
      .sort((left, right) => left.name.localeCompare(right.name));
  });
  const status = computed(() => (device.value ? statusOf(device.value) : 'missing'));
  const values = computed(() => derived.values.value);
</script>

<template>
  <section class="nh-presence-card" :aria-label="`Presence for ${device?.friendly_name ?? card.device_id}`">
    <header>
      <div>
        <strong>{{ device?.friendly_name ?? 'Unknown device' }}</strong>
        <span class="nh-muted"> · {{ status }}</span>
      </div>
      <UiButton v-if="editing" size="sm" variant="ghost" icon="delete" label="Remove presence card" :disabled="busy" @click="$emit('remove')" />
    </header>
    <p v-if="!device" class="nh-muted">This presence device is unavailable. Remove the card if it is no longer needed.</p>
    <template v-else>
      <p v-if="!derived.derived.value" class="nh-muted">Presence has not been derived for this device yet.</p>
      <template v-else>
        <p class="nh-presence-state">
          <strong>{{ values.occupied === null ? 'Unknown' : values.occupied ? 'Occupied' : 'Clear' }}</strong>
          <span v-if="values.targetCount !== null"> · {{ values.targetCount }} target{{ values.targetCount === 1 ? '' : 's' }}</span>
          <span v-if="values.lastZone"> · last zone {{ values.lastZone }}</span>
          <span v-if="values.emptySeconds !== null && !values.occupied"> · empty {{ values.emptySeconds }} s</span>
        </p>
        <ul v-if="zones.length" class="nh-presence-zones" aria-label="Zone occupancy">
          <li v-for="zone in zones" :key="zone.id" :class="{ 'is-occupied': zone.occupied }">
            {{ zone.name }}: {{ zone.occupied === null ? 'unknown' : zone.occupied ? 'occupied' : 'clear' }}
          </li>
        </ul>
        <RouterLink :to="{ name: RouteName.DevicePage, params: { id: card.device_id } }">Open device</RouterLink>
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
