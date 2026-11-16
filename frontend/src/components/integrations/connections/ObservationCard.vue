<script setup lang="ts">
  /*
   * Catalog card for a device-backed integration.
   *
   * The rest of the Integrations page lists configured *instances* — a weather location, a bin
   * collection address. Wireless observation has no instance to configure: it becomes available
   * when a capable device is enrolled, and the "configuration" is that enrolment. So it gets its
   * own card rather than a row in the instance board, which would need a fake instance record.
   *
   * It still belongs on this page. The maintainer's objection was that a top-level "Wireless" nav
   * item made the hub look like it had a wireless feature; here it reads as what it is — one more
   * thing the hub can do, alongside the others, driven by hardware you own.
   */
  import { computed } from 'vue';
  import { useRouter } from 'vue-router';
  import type { Device } from '@/types/device';
  import { observationDevices, OBSERVATION_ROUTE } from '@/domain/integrations/device-backed';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';

  const props = defineProps<{
    /** Enrolled devices; the card renders only when at least one can observe. */
    devices: readonly Device[];
  }>();

  const router = useRouter();
  const sources = computed(() => observationDevices(props.devices));

  const summary = computed(() => {
    const count = sources.value.length;
    if (count === 1) return `Reading through ${sources.value[0].friendly_name}.`;
    return `Reading through ${count} sensors. Devices heard by more than one are shown once.`;
  });

  function openRadar(): void {
    void router.push({ name: OBSERVATION_ROUTE });
  }
</script>

<template>
  <article v-if="sources.length > 0" class="observation-card" aria-labelledby="observation-card-title">
    <span class="observation-card__icon" aria-hidden="true">
      <UiIcon name="signal" />
    </span>

    <div class="observation-card__body">
      <h2 id="observation-card-title" class="observation-card__title">Wireless observation</h2>
      <p class="observation-card__summary">{{ summary }}</p>

      <!--
        Sources are named rather than counted only. The operator chose these devices, and seeing
        which ones are contributing is how they notice a sensor that stopped reporting.
      -->
      <ul class="observation-card__sources">
        <li v-for="source in sources" :key="source.id">{{ source.friendly_name }}</li>
      </ul>
    </div>

    <UiButton variant="primary" @click="openRadar">Open radar</UiButton>
  </article>
</template>

<style scoped>
  .observation-card {
    display: flex;
    align-items: center;
    gap: 0.85rem;
    padding: 0.85rem 1rem;
    border: 1px solid var(--border-subtle, #dee2e6);
    border-radius: 0.6rem;
    background: var(--surface-raised, rgba(255, 255, 255, 0.02));
  }

  .observation-card__icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 2.25rem;
    height: 2.25rem;
    border-radius: 0.5rem;
    background: var(--accent-soft, rgba(74, 158, 255, 0.15));
    color: var(--accent-strong, #4a9eff);
    flex: 0 0 auto;
  }

  .observation-card__body {
    flex: 1 1 auto;
    min-width: 0;
  }

  .observation-card__title {
    margin: 0;
    font-size: 0.95rem;
  }

  .observation-card__summary {
    margin: 0.15rem 0 0;
    font-size: 0.78rem;
    color: var(--text-muted, #8b9bab);
  }

  .observation-card__sources {
    display: flex;
    flex-wrap: wrap;
    gap: 0.25rem 0.6rem;
    margin: 0.3rem 0 0;
    padding: 0;
    list-style: none;
    font-size: 0.72rem;
    color: var(--text-subtle, #7c8b9c);
  }

  /* Phone: the action drops below the text rather than squeezing it. */
  @media (max-width: 600px) {
    .observation-card {
      flex-wrap: wrap;
    }

    .observation-card__body {
      flex: 1 1 100%;
      order: 2;
    }

    .observation-card__icon {
      order: 1;
    }

    .observation-card > :last-child {
      order: 3;
      width: 100%;
    }
  }
</style>
