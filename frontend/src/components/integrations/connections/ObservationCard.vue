<script setup lang="ts">
  /*
   * Catalog entry for a device-backed integration.
   *
   * The rest of the Integrations page lists configured *instances* — a weather location, a bin
   * collection address. Wireless observation has no instance: it becomes available when a capable
   * device is enrolled, and the "configuration" is that enrolment. So it renders as an entry rather
   * than an instance row, reusing the same classes so it sits in the same visual rhythm instead of
   * looking like a different kind of object.
   *
   * Where the other entries carry a schedule and a refresh action, this carries its sources. That is
   * the equivalent information: what it reads from, and whether those are reporting.
   */
  import { computed } from 'vue';
  import { useRouter } from 'vue-router';
  import type { Device } from '@/types/device';
  import { observationDevices, OBSERVATION_ROUTE } from '@/domain/integrations/device-backed';
  import UiIcon from '@/components/ui/UiIcon.vue';

  const props = defineProps<{
    /** Enrolled devices; the card renders only when at least one can observe. */
    devices: readonly Device[];
  }>();

  const router = useRouter();
  const sources = computed(() => observationDevices(props.devices));

  /** Devices declaring the capability that are currently offline, so a stale source is visible. */
  const offline = computed(() =>
    sources.value.filter((device) => device.availability !== 'online').map((device) => device.friendly_name),
  );

  const summary = computed(() => {
    const count = sources.value.length;
    const noun = count === 1 ? 'sensor' : 'sensors';
    return `Reading through ${count} ${noun}`;
  });

  function openRadar(): void {
    void router.push({ name: OBSERVATION_ROUTE });
  }
</script>

<template>
  <!--
    Shares the entry *styles* but carries its own class for selection. Existing tests and the board
    treat `.integration-entry` as "a configured instance", so a second element with that class
    silently changed counts and broke six specs. Styling and identity are different concerns.
  -->
  <article v-if="sources.length > 0" class="observation-entry" aria-label="Wireless observation">
    <span class="observation-entry-icon" aria-hidden="true">
      <UiIcon name="signal" />
    </span>

    <div class="observation-entry-main">
      <div class="observation-entry-heading">
        <h3>
          <RouterLink :to="{ name: OBSERVATION_ROUTE }" class="observation-entry-link">
            Wireless observation
          </RouterLink>
        </h3>
        <span class="observation-entry-type">Bluetooth</span>
      </div>

      <p class="observation-entry-summary">
        <strong>{{ summary }}</strong>
        <span>{{ sources.map((source) => source.friendly_name).join(', ') }}</span>
      </p>

      <!--
        No refresh action: there is nothing to fetch. The equivalent line to another entry's
        schedule says what this one does instead, so the absence does not read as a missing control.
      -->
      <p class="observation-entry-schedule">Updated continuously while a sensor is connected</p>

      <p v-if="offline.length > 0" class="observation-entry-status is-warning">
        Not reporting: {{ offline.join(', ') }}
      </p>
      <p v-else class="observation-entry-status is-ok">All sensors reporting</p>
    </div>

    <div class="observation-entry-actions">
      <button type="button" class="observation-entry-open" @click="openRadar">Open radar</button>
    </div>
  </article>
</template>

<style scoped>
  /*
   * Its own class family, matching the instance entries' proportions rather than sharing their
   * class names.
   *
   * Sharing `.integration-entry` would have been less code, but that class means "a configured
   * instance" to the board and to six browser tests — a second element carrying it silently changed
   * their counts. Styling and identity are different concerns, so the values are duplicated here on
   * purpose: they are a handful of layout rules, and the alternative was breaking the meaning of a
   * selector used elsewhere.
   */
  .observation-entry {
    display: flex;
    align-items: flex-start;
    gap: 0.8rem;
    min-width: 0;
    padding: 0.9rem 1rem;
    border: var(--nh-border-w) solid var(--nh-border, #2a3441);
    border-radius: var(--nh-radius-l, 0.6rem);
    background: var(--nh-surface, transparent);
  }

  .observation-entry-main {
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
    min-width: 0;
    flex: 1;
  }

  .observation-entry-heading {
    display: flex;
    align-items: baseline;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .observation-entry-heading h3 {
    margin: 0;
    font-size: 0.95rem;
    overflow-wrap: anywhere;
  }

  .observation-entry-link {
    text-decoration: none;
  }

  .observation-entry-link:hover {
    text-decoration: underline;
  }

  .observation-entry-type {
    color: var(--nh-text-muted, #8b9bab);
    font-size: 0.72rem;
  }

  .observation-entry-summary {
    display: flex;
    flex-wrap: wrap;
    gap: 0.35rem 0.6rem;
    margin: 0.15rem 0 0;
    font-size: 0.84rem;
    overflow-wrap: anywhere;
  }

  .observation-entry-summary strong {
    font-weight: 600;
  }

  .observation-entry-schedule {
    margin: 0.15rem 0 0;
    color: var(--nh-text-muted, #8b9bab);
    font-size: 0.75rem;
  }

  .observation-entry-status {
    font-size: 0.75rem;
    margin-top: 0.2rem;
  }

  .observation-entry-status.is-warning {
    color: var(--nh-warning, #d9a441);
  }

  .observation-entry-status.is-ok {
    color: var(--nh-text-muted, #8b9bab);
  }

  .observation-entry-actions {
    display: flex;
    align-items: flex-start;
    gap: 0.15rem;
    flex: none;
  }

  .observation-entry-icon {
    display: grid;
    place-items: center;
    flex: none;
    width: 2.5rem;
    height: 2.5rem;
    border-radius: 50%;
    background: var(--nh-surface-2, rgba(74, 158, 255, 0.14));
    color: var(--nh-accent, #4a9eff);
  }

  .observation-entry-icon :deep(svg) {
    width: 1.3rem;
    height: 1.3rem;
  }

  /* A text action rather than an icon button: "Open radar" is a destination, not a command, and the
     instance entries use icons only for immediate actions like refresh. */
  .observation-entry-open {
    padding: 0.35rem 0.7rem;
    border: var(--nh-border-w, 1px) solid var(--nh-border, #2a3441);
    border-radius: 0.4rem;
    background: transparent;
    color: var(--nh-text, inherit);
    font: inherit;
    font-size: 0.8rem;
    cursor: pointer;
    white-space: nowrap;
  }

  .observation-entry-open:hover {
    border-color: var(--nh-accent, #4a9eff);
    color: var(--nh-accent, #4a9eff);
  }

  .observation-entry-open:focus-visible {
    outline: 2px solid var(--nh-accent, #4a9eff);
    outline-offset: 2px;
  }
</style>
