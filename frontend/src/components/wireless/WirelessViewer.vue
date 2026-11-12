<script setup lang="ts">
  /*
   * Wireless observation viewer (spec 064).
   *
   * A hub-level live instrument, not configuration: it shows whatever any enrolled observing
   * source reports. Nothing here names a provider, so a new source kind appears without a
   * component change.
   */
  import { computed, ref } from 'vue';
  import { useStore } from 'vuex';
  import type { ObservationDecision, ObservationIdentityView } from '@/types/observations.type';
  import { ObservationDecisions, ObservationFacets } from '@/types/observations.type';
  import { snapshotSummary, sourceIsSilent, sourceStatus, untriagedCount, vendorHex } from '@/domain/observations/presentation';
  import ObservationRadar from './ObservationRadar.vue';
  import ObservationTable from './ObservationTable.vue';
  import ObservationFilters from './ObservationFilters.vue';

  const store = useStore();

  const snapshot = computed(() => store.getters['observations/snapshot']);
  const loaded = computed(() => store.getters['observations/hasData']);
  const stale = computed(() => store.getters['observations/stale']);
  const allIdentities = computed<ObservationIdentityView[]>(() => snapshot.value?.identities ?? []);

  const selected = ref<Record<string, string[]>>({});
  const onlyNew = ref(false);
  /** Keys with a save in flight, so the row reports progress rather than looking inert. */
  const pending = ref<Set<string>>(new Set());

  /** Facets declared by any source; the filter component renders only these. */
  const facets = computed(() => {
    const declared = new Set<string>();
    for (const source of snapshot.value?.sources ?? []) {
      for (const facet of source.facets) declared.add(facet);
    }
    return [...declared];
  });

  /** Apply the operator's facet selection and the new-only toggle. */
  const visible = computed(() => {
    const selection = selected.value;
    return allIdentities.value.filter((identity) => {
      if (onlyNew.value && identity.decision !== ObservationDecisions.Unknown) return false;
      return Object.entries(selection).every(([facet, values]) => {
        if (values.length === 0) return true;
        if (facet === ObservationFacets.Kind) return values.includes(identity.kind);
        if (facet === ObservationFacets.Vendor) return values.includes(vendorHex(identity.vendor));
        // An unknown facet cannot be filtered here; ignoring it keeps a newer backend usable.
        return true;
      });
    });
  });

  const summary = computed(() => snapshotSummary(snapshot.value));
  const newCount = computed(() => untriagedCount(snapshot.value));
  const truncated = computed(() => snapshot.value?.truncated === true);
  /** Sources that are connected but have delivered nothing, which usually means a config fault. */
  const sources = computed(() => snapshot.value?.sources ?? []);
  const silentSources = computed(() => sources.value.filter(sourceIsSilent));

  /**
   * Persist a triage decision. The row is marked pending until the request settles, and the
   * snapshot is not updated locally: the backend is the authority and broadcasts the result.
   */
  function decide(key: string, decision: ObservationDecision): void {
    pending.value = new Set(pending.value).add(key);
    void store
      .dispatch('ws/emit', {
        event: 'saveObservationIdentity',
        message: { key, decision },
      })
      .finally(() => {
        const next = new Set(pending.value);
        next.delete(key);
        pending.value = next;
      });
  }
</script>

<template>
  <section class="wireless-viewer" aria-labelledby="wireless-title">
    <header class="wireless-viewer__header">
      <div>
        <h1 id="wireless-title">Wireless</h1>
        <p class="wireless-viewer__summary" aria-live="polite">
          {{ summary }}
          <span v-if="newCount > 0">· {{ newCount }} untriaged</span>
        </p>
      </div>
      <p v-if="stale" class="wireless-viewer__stale" role="status">
        Disconnected. Showing the last received data.
      </p>
    </header>

    <p v-if="!loaded" class="wireless-viewer__empty">
      Waiting for the first observation. Enrol a device with Bluetooth proxy enabled to begin.
    </p>

    <template v-else>
      <p v-if="truncated" class="wireless-viewer__notice" role="status">
        Showing the most recently heard devices only; the snapshot reached its size limit.
      </p>

      <!--
        Silence has several indistinguishable causes, so each source states its own delivery
        status. Without this an empty viewer could mean an unreachable host, a device without
        bluetooth_proxy, or genuinely nothing in range.
      -->
      <ul v-if="sources.length > 0" class="wireless-viewer__sources">
        <li v-for="source in sources" :key="source.id" :class="{ 'is-silent': sourceIsSilent(source) }">
          <strong>{{ source.name }}</strong>
          <span>{{ sourceStatus(source) }}</span>
        </li>
      </ul>

      <p v-if="silentSources.length > 0" class="wireless-viewer__notice" role="status">
        {{ silentSources.length === 1 ? 'One source is' : `${silentSources.length} sources are` }}
        connected but not sending advertisements. Confirm each device has
        <code>bluetooth_proxy</code> in its ESPHome configuration.
      </p>

      <div class="wireless-viewer__layout">
        <ObservationRadar :identities="visible" />
        <div class="wireless-viewer__detail">
          <ObservationFilters
            :identities="allIdentities"
            :facets="facets"
            :selected="selected"
            :only-new="onlyNew"
            @update:selected="selected = $event"
            @update:only-new="onlyNew = $event"
          />
          <ObservationTable :identities="visible" :pending="pending" @decide="decide" />
        </div>
      </div>
    </template>
  </section>
</template>

<style scoped>
  .wireless-viewer {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    padding: 1.5rem;
  }

  .wireless-viewer__header {
    display: flex;
    justify-content: space-between;
    gap: 1rem;
    flex-wrap: wrap;
    align-items: baseline;
  }

  .wireless-viewer__header h1 {
    font-size: 1.375rem;
    margin: 0;
  }

  .wireless-viewer__summary,
  .wireless-viewer__empty,
  .wireless-viewer__notice {
    color: var(--text-muted, #8b9bab);
    margin: 0.25rem 0 0;
    font-size: 0.875rem;
  }

  .wireless-viewer__stale {
    color: var(--warning-text, #e0a33e);
    margin: 0;
    font-size: 0.875rem;
  }

  .wireless-viewer__sources {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-wrap: wrap;
    gap: 0.75rem;
  }

  .wireless-viewer__sources li {
    display: flex;
    flex-direction: column;
    gap: 0.125rem;
    padding: 0.375rem 0.625rem;
    border: 1px solid var(--border-subtle, #2a3441);
    border-radius: 0.375rem;
    font-size: 0.8125rem;
  }

  /* A silent source is a likely misconfiguration, so it is marked by border and text, not
     colour alone. */
  .wireless-viewer__sources li.is-silent {
    border-style: dashed;
    border-color: var(--warning-text, #e0a33e);
  }

  .wireless-viewer__sources li span {
    color: var(--text-muted, #8b9bab);
  }

  .wireless-viewer__layout {
    display: grid;
    gap: 1.5rem;
    grid-template-columns: minmax(16rem, 24rem) minmax(0, 1fr);
    align-items: start;
  }

  .wireless-viewer__detail {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    min-width: 0;
  }

  /* Stack the radar above the table on narrow screens rather than shrinking both. */
  @media (max-width: 899px) {
    .wireless-viewer {
      padding: 1rem;
    }

    .wireless-viewer__layout {
      grid-template-columns: 1fr;
    }
  }
</style>
