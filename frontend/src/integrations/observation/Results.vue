<script setup lang="ts">
  /*
   * What the sensor is hearing right now, on the integration page.
   *
   * It is resolved on read (ADR-005), so these numbers are current rather than a snapshot of the
   * moment the integration was added. That is what makes this page worth having: it answers "what is
   * it hearing?" without leaving for the radar, and it cannot disagree with the radar, the board row
   * or a Home tile because they all read the same live snapshot.
   *
   * Packet counters are deliberately demoted. They answer "is the radio working", not "what is
   * here": they are lifetime totals for this run, they reset on restart, and pairing them beside a
   * device count invites reading them as results. They live under a disclosure, where an operator
   * diagnosing a silent sensor will find them and nobody else has to.
   */
  import { computed } from 'vue';
  import type { ObservationCensus } from '@/types/integrations';
  import { RouteName } from '@/types/router';
  import { censusReporting, censusSummary } from '@/domain/integrations/observation';

  const props = defineProps<{ resource: ObservationCensus }>();

  /**
   * Headline and status note, computed together in the domain module.
   *
   * They are one call rather than two independent computeds because separating them is what produced
   * the contradiction: the count was phrased in the present tense while nothing was reporting.
   */
  const summary = computed(() => censusSummary(props.resource));

  /** Whether any source is hearing anything, which decides what the packet disclosure is for. */
  const reporting = computed(() => censusReporting(props.resource));

  /** Counts worth showing only when they explain something: silence, or a drop. */
  const noteworthy = computed(
    () => !reporting.value || props.resource.sources.some((source) => source.dropped > 0)
  );
</script>

<template>
  <section class="observation-census" aria-label="Wireless observation now">
    <p class="observation-census-summary">
      <!--
        The count and its unit are one string, not a <strong> followed by a text node. Vue condenses
        the whitespace between them — it rendered "23devices in range" — and a number jammed against
        its noun reads as a typo in exactly the place the operator is looking.
      -->
      <strong>{{ summary.headline }}</strong>
    </p>
    <p v-if="resource.hidden > 0" class="observation-census-note">
      {{ resource.hidden }} {{ resource.hidden === 1 ? 'device is' : 'devices are' }} hidden by the class filter.
    </p>
    <!--
      One status note, not a per-condition pile. The contradiction came from stacking independent
      sentences: each described a different subset of the sources, and together they asserted both
      "devices are here" and "nothing is reporting" about the same moment.
    -->
    <p v-if="summary.note" class="observation-census-note" role="status">{{ summary.note }}</p>
    <ul v-if="resource.sources.length" class="observation-census-sources">
      <li v-for="source in resource.sources" :key="source.id">{{ source.name }}</li>
    </ul>
    <p class="observation-census-note">
      <RouterLink :to="{ name: RouteName.Wireless }">Open the radar</RouterLink> for who is where, what
      they are, and the full history.
    </p>
    <details v-if="resource.sources.length" class="observation-census-packets" :open="noteworthy">
      <summary>Packets</summary>
      <ul class="observation-census-sources">
        <li v-for="source in resource.sources" :key="source.id">
          <span>{{ source.name }}</span>
          <span class="observation-census-counts">
            {{ source.advertisements }} heard · {{ source.decoded }} decoded
            <template v-if="source.dropped > 0"> · {{ source.dropped }} dropped</template>
          </span>
        </li>
      </ul>
      <p class="observation-census-note">
        Totals since the hub started, so they reset when it restarts. They say whether the radio is
        delivering, not how many devices are nearby.
      </p>
    </details>
  </section>
</template>

<style scoped>
  .observation-census {
    display: grid;
    gap: 0.5rem;
  }
  .observation-census-summary {
    margin: 0;
    font-size: 1rem;
  }
  .observation-census-sources {
    display: grid;
    gap: 0.25rem;
    margin: 0;
    padding: 0;
    list-style: none;
  }
  .observation-census-sources li {
    display: flex;
    flex-wrap: wrap;
    gap: 0.4rem;
    justify-content: space-between;
    padding: 0.35rem 0;
    border-bottom: 1px solid var(--nh-border, #2a3441);
  }
  .observation-census-counts {
    color: var(--nh-text-muted, #8b9bab);
    font-size: 0.82rem;
  }
  .observation-census-note {
    margin: 0;
    color: var(--nh-text-muted, #8b9bab);
    font-size: 0.82rem;
  }
  .observation-census-packets summary {
    cursor: pointer;
    color: var(--nh-text-muted, #8b9bab);
    font-size: 0.82rem;
  }
  .observation-census-packets[open] summary {
    margin-bottom: 0.4rem;
  }
</style>
