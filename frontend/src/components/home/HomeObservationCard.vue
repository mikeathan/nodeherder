<script setup lang="ts">
  /*
   * Home shortcut for wireless observation.
   *
   * A compact summary rather than a live radar: a canvas inside a tile grid is cramped, and the
   * radar needs room to be readable. This says how many sensors are reporting and how many devices
   * have been heard, and opens the real view.
   *
   * It exists because the radar is no longer a nav item. Without a shortcut the feature would be
   * reachable only through Integrations, which is a place you visit to configure things, not to
   * glance at them.
   *
   * The tile assumes the integration has been added. Adding it is a separate action, which is what
   * makes this removable: a tile that could not be removed would be exactly the behaviour this
   * replaced.
   */
  import { computed } from 'vue';
  import { useRouter } from 'vue-router';
  import { store } from '@/store';
  import { useHub } from '@/composables/useHub';
  import { observationDevices, OBSERVATION_ROUTE } from '@/domain/integrations/observation';
  import { snapshotSummary } from '@/domain/observations/presentation';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';

  const props = defineProps<{
    editing?: boolean;
    busy?: boolean;
  }>();

  const emit = defineEmits<{ remove: [] }>();
  const router = useRouter();

  // `useHub` is the established way for a home tile to read devices; the store is still read
  // directly for the observation snapshot because there is no composable for it yet.
  const { devices } = useHub();
  const sensors = computed(() => observationDevices(devices.value));
  const snapshot = computed(() => store.getters['observations/snapshot']);

  /**
   * What the radar is currently seeing, or why it is not seeing anything.
   *
   * The two states need different wording: "no devices heard" is normal early on, while "no sensor"
   * means nothing is enrolled and the card should not pretend otherwise. Not observing is a third
   * state, and it is the one that used to be unreachable: the snapshot outlived the integration, so
   * a removed integration still showed its last census forever.
   */
  const summary = computed(() => {
    if (sensors.value.length === 0) return 'No sensor enrolled';
    if (!snapshot.value) return 'Not observing — add it as an integration to start';
    if (snapshot.value.observing === false) return 'Not observing';
    return snapshotSummary(snapshot.value);
  });

  function open(): void {
    void router.push({ name: OBSERVATION_ROUTE });
  }
</script>

<template>
  <article class="nh-observation-tile" :data-home-card="`observation`" :aria-label="`Wireless observation: ${summary}`">
    <header class="nh-observation-head">
      <span class="nh-observation-icon" aria-hidden="true"><UiIcon name="signal" /></span>
      <h3 class="nh-observation-title">
        <!--
          The whole tile is the link, via an overlay, exactly as IntegrationTile does. A card
          containing one button made the natural target inert and put a primary-coloured button on
          a dashboard where no neighbouring tile has one.
        -->
        <RouterLink v-if="!editing" class="nh-observation-link" :to="{ name: OBSERVATION_ROUTE }">
          Wireless observation
        </RouterLink>
        <template v-else>Wireless observation</template>
      </h3>
    </header>

    <p class="nh-observation-summary">{{ summary }}</p>
    <p v-if="sensors.length > 0" class="nh-observation-sensors">
      Via {{ sensors.map((sensor) => sensor.friendly_name).join(', ') }}
    </p>

    <div v-if="editing" class="nh-observation-actions">
      <UiButton size="sm" variant="ghost" :disabled="props.busy" @click="open">View radar</UiButton>
      <UiButton size="sm" variant="ghost" :disabled="props.busy" @click="emit('remove')">Remove</UiButton>
    </div>
  </article>
</template>

<style scoped>
  /*
   * Tokens only, and only ones the theme layer defines.
   *
   * This tile previously styled with --border-subtle, --surface-raised, --accent-strong and
   * --text-muted. None of the four exist here, so the hardcoded fallbacks won on every theme and
   * the border rendered as a near-white outline on a dark dashboard. Using the same tokens as
   * IntegrationTile is what makes it read as one of the family rather than as an alert.
   */
  .nh-observation-tile {
    position: relative;
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
    padding: 0.7rem 0.75rem;
    inline-size: 100%;
    height: 100%;
    border: var(--nh-border-w) solid var(--nh-border);
    border-radius: var(--nh-radius-l);
    background: var(--nh-surface);
    box-shadow: var(--nh-shadow-1);
  }

  .nh-observation-head {
    display: flex;
    align-items: center;
    gap: 0.4rem;
  }

  .nh-observation-icon {
    display: grid;
    place-items: center;
    flex: none;
    color: var(--nh-accent);
  }

  .nh-observation-title {
    margin: 0;
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--nh-text);
  }

  .nh-observation-link {
    color: inherit;
    text-decoration: none;
  }

  /* The overlay that makes the whole tile clickable, as IntegrationTile's link does. */
  .nh-observation-link::after {
    content: '';
    position: absolute;
    inset: 0;
    border-radius: var(--nh-radius-l);
  }

  .nh-observation-link:focus-visible {
    outline: none;
  }

  .nh-observation-link:focus-visible::after {
    outline: 3px solid var(--nh-accent);
    outline-offset: 3px;
  }

  .nh-observation-summary {
    margin: 0;
    font-size: 0.82rem;
    color: var(--nh-text-muted);
    overflow-wrap: anywhere;
  }

  .nh-observation-sensors {
    margin: 0;
    font-size: 0.7rem;
    color: var(--nh-text-muted);
    overflow-wrap: anywhere;
  }

  /* Pushed to the bottom so tiles of different content still align their actions. */
  .nh-observation-actions {
    display: flex;
    gap: 0.35rem;
    margin-top: auto;
    padding-top: 0.4rem;
  }
</style>
