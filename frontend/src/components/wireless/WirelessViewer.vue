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
  import { ObservationDecisions } from '@/types/observations.type';
  import { snapshotSummary, untriagedCount } from '@/domain/observations/presentation';
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
        if (facet === 'kind') return values.includes(identity.kind);
        if (facet === 'vendor') {
          const hex = identity.vendor === undefined ? '' : `0x${identity.vendor.toString(16).toUpperCase().padStart(4, '0')}`;
          return values.includes(hex);
        }
        // An unknown facet cannot be filtered here; ignoring it keeps a newer backend usable.
        return true;
      });
    });
  });

  const summary = computed(() => snapshotSummary(snapshot.value));
  const newCount = computed(() => untriagedCount(snapshot.value));
  const truncated = computed(() => snapshot.value?.truncated === true);

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
