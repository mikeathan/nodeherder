<script setup lang="ts">
  /*
   * Identity table for the wireless viewer (spec 064).
   *
   * This is where per-packet detail and triage live. The canvas shows aggregated state; the
   * table shows what the operator acts on, including the rotating-address marker which stops a
   * re-keyed phone from looking like a new arrival.
   */
  import { computed } from 'vue';
  import type { ObservationDecision, ObservationIdentityView } from '@/types/observations.type';
  import { ObservationDecisions } from '@/types/observations.type';
  import { identityTitle, isEphemeralIdentity, kindLabel } from '@/domain/observations/presentation';
  import ClassBadge from './ClassBadge.vue';
  import UiButton from '@/components/ui/UiButton.vue';

  const props = defineProps<{
    identities: readonly ObservationIdentityView[];
    /** Identity keys with a save in flight, so their controls report progress. */
    pending: ReadonlySet<string>;
    /** Currently isolated identity, so the row can be marked. */
    selectedKey?: string | null;
  }>();

  const emit = defineEmits<{
    (event: 'decide', key: string, decision: ObservationDecision): void;
    (event: 'select', key: string): void;
  }>();

  function time(value: string): string {
    const parsed = new Date(value);
    if (Number.isNaN(parsed.getTime())) return '—';
    return parsed.toLocaleTimeString();
  }

  function signal(rssi: number | null | undefined): string {
    // An absent reading must not render as 0 dBm, which would read as a strong signal.
    if (typeof rssi !== 'number' || !Number.isFinite(rssi)) return '—';
    return `${rssi} dBm`;
  }

  function addressType(identity: ObservationIdentityView): string {
    if (isEphemeralIdentity(identity)) return 'Rotating';
    return 'Stable';
  }

  const rows = computed(() => props.identities);

  function isPending(key: string): boolean {
    return props.pending.has(key);
  }
</script>

<template>
  <div class="observation-table">
    <table>
      <caption class="observation-table__caption">
        Every identity any observing source has heard, most recent first.
      </caption>
      <thead>
        <tr>
          <th scope="col">Identity</th>
          <th scope="col">Class</th>
          <th scope="col">Kind</th>
          <th scope="col">Address</th>
          <th scope="col">Signal</th>
          <th scope="col">Packets</th>
          <th scope="col">Last seen</th>
          <th scope="col">Triage</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="identity in rows"
          :key="identity.key"
          :class="{ 'is-selected': props.selectedKey === identity.key }"
        >
          <th scope="row">
            <!--
              The identity title is the isolate control. It is a button rather than a row click
              so the triage buttons beside it stay independently operable.
            -->
            <button
              type="button"
              class="observation-table__title observation-table__select"
              :aria-pressed="props.selectedKey === identity.key"
              :aria-label="`Isolate ${identityTitle(identity)}`"
              title="Isolate this identity"
              @click="emit('select', identity.key)"
            >
              <span class="observation-table__target-icon" aria-hidden="true">◎</span>
              {{ identityTitle(identity) }}
            </button>
            <span
              v-if="identity.decision === ObservationDecisions.Unknown"
              class="observation-table__new"
            >
              New
            </span>
          </th>
          <td>
            <!-- The class is advisory, so it sits beside the triage controls rather than replacing
                 them: the operator's decision is the authority. -->
            <ClassBadge :classification="identity.classification" compact />
          </td>
          <!-- Rendered through kindLabel: the wire value is "adv", which is jargon for what is
               simply Bluetooth. -->
          <td>{{ kindLabel(identity.kind) }}</td>
          <td>
            <span :title="identity.key">{{ addressType(identity) }}</span>
          </td>
          <td>{{ signal(identity.rssi) }}</td>
          <td>{{ identity.packets }}</td>
          <td>{{ time(identity.lastSeen) }}</td>
          <td>
            <div class="observation-table__actions" role="group" :aria-label="`Triage for ${identityTitle(identity)}`">
              <UiButton
                v-if="identity.decision !== ObservationDecisions.Known"
                :disabled="isPending(identity.key)"
                size="sm"
                variant="default"
                @click="emit('decide', identity.key, ObservationDecisions.Known)"
              >
                Mark known
              </UiButton>
              <UiButton
                v-if="identity.decision !== ObservationDecisions.Ignored"
                :disabled="isPending(identity.key)"
                size="sm"
                variant="ghost"
                @click="emit('decide', identity.key, ObservationDecisions.Ignored)"
              >
                Ignore
              </UiButton>
              <span v-if="isPending(identity.key)" class="observation-table__pending" role="status">
                Saving…
              </span>
            </div>
          </td>
        </tr>
      </tbody>
    </table>

    <p v-if="rows.length === 0" class="observation-table__empty">
      No identities yet. Advertisements appear here as sources report them.
    </p>
  </div>
</template>

<style scoped>
  .observation-table {
    overflow-x: auto;
  }

  .observation-table__caption {
    text-align: left;
    color: var(--text-muted, #8b9bab);
    font-size: 0.8125rem;
    padding-bottom: 0.5rem;
  }

  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.875rem;
  }

  th,
  td {
    text-align: left;
    padding: 0.5rem 0.75rem;
    border-bottom: 1px solid var(--border-subtle, #2a3441);
    white-space: nowrap;
  }

  thead th {
    color: var(--text-muted, #8b9bab);
    font-weight: 600;
  }

  .observation-table__title {
    font-weight: 600;
  }

  /* The isolate control looks like text until focused, so the table stays scannable. */
  /* Visibly a control: without the resting underline and target glyph, isolation was
     undiscoverable and the viewer looked unchanged. */
  .observation-table__select {
    display: inline-flex;
    align-items: center;
    gap: 0.3rem;
    background: none;
    border: 0;
    color: inherit;
    font: inherit;
    padding: 0;
    text-align: left;
    cursor: pointer;
    text-decoration: underline;
    text-decoration-color: color-mix(in srgb, currentColor 45%, transparent);
    text-underline-offset: 0.15em;
  }

  .observation-table__select:hover,
  .observation-table__select:focus-visible {
    text-decoration-color: currentColor;
    color: var(--accent-primary, #4a9eff);
  }

  .observation-table__target-icon {
    color: var(--accent-primary, #4a9eff);
  }

  .observation-table__select:focus-visible {
    outline: 2px solid var(--accent-primary, #4a9eff);
    outline-offset: 2px;
  }

  tr.is-selected {
    background: color-mix(in srgb, var(--accent-primary, #4a9eff) 12%, transparent);
  }

  /* The new marker carries text, not just colour, so it survives grayscale and screen readers. */
  .observation-table__new {
    margin-left: 0.5rem;
    padding: 0 0.375rem;
    border-radius: 999px;
    background: var(--accent-primary, #4a9eff);
    color: #04121f;
    font-size: 0.6875rem;
    font-weight: 700;
    text-transform: uppercase;
  }

  .observation-table__actions {
    display: flex;
    gap: 0.375rem;
    align-items: center;
  }

  .observation-table__pending {
    color: var(--text-muted, #8b9bab);
    font-size: 0.8125rem;
  }

  .observation-table__empty {
    color: var(--text-muted, #8b9bab);
    padding: 1rem 0.75rem;
  }
</style>
