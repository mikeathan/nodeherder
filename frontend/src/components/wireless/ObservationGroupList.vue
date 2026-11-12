<script setup lang="ts">
  /*
   * Grouped list view for wireless observations (spec 064).
   *
   * A dense, scannable alternative to the radar. Real households show dozens of identities, so
   * rows are sectioned by kind or vendor: ungrouped, the list is a wall of opaque keys. Each
   * section is collapsible so a noisy vendor can be folded away without ignoring its devices.
   */
  import { ref } from 'vue';
  import type { ObservationDecision, ObservationIdentityView } from '@/types/observations.type';
  import { identityTitle, isEphemeralIdentity } from '@/domain/observations/presentation';
  import UiButton from '@/components/ui/UiButton.vue';

  const props = defineProps<{
    groups: readonly { group: string; identities: ObservationIdentityView[] }[];
    /** Identity keys with a save in flight, so their controls report progress. */
    pending: ReadonlySet<string>;
  }>();

  const emit = defineEmits<{
    (event: 'decide', key: string, decision: ObservationDecision): void;
  }>();

  /** Section names the operator has folded. Collapsed state is display-only. */
  const collapsed = ref<Set<string>>(new Set());

  function isCollapsed(group: string): boolean {
    return collapsed.value.has(group);
  }

  function toggleGroup(group: string): void {
    const next = new Set(collapsed.value);
    if (next.has(group)) next.delete(group);
    else next.add(group);
    collapsed.value = next;
  }

  function time(value: string): string {
    const parsed = new Date(value);
    if (Number.isNaN(parsed.getTime())) return '—';
    return parsed.toLocaleTimeString();
  }

  /** Absent readings must not render as 0 dBm, which would read as a strong signal. */
  function signal(rssi: number | null | undefined): string {
    if (typeof rssi !== 'number' || !Number.isFinite(rssi)) return '—';
    return `${rssi} dBm`;
  }
</script>

<template>
  <div class="observation-groups">
    <section v-for="group in props.groups" :key="group.group" class="observation-groups__section">
      <h2 class="observation-groups__heading">
        <button
          type="button"
          class="observation-groups__toggle"
          :aria-expanded="!isCollapsed(group.group)"
          @click="toggleGroup(group.group)"
        >
          <span class="observation-groups__chevron" aria-hidden="true">{{ isCollapsed(group.group) ? '▸' : '▾' }}</span>
          {{ group.group }}
          <span class="observation-groups__count">{{ group.identities.length }}</span>
        </button>
      </h2>

      <ul v-show="!isCollapsed(group.group)" class="observation-groups__list">
        <li v-for="identity in group.identities" :key="identity.key" class="observation-groups__row">
          <span class="observation-groups__title">
            {{ identityTitle(identity) }}
            <span v-if="identity.decision === 'unknown'" class="observation-groups__new">New</span>
          </span>
          <span class="observation-groups__meta">
            <span v-if="isEphemeralIdentity(identity)" title="Rotating address">rotating</span>
            <span>{{ signal(identity.rssi) }}</span>
            <span>{{ identity.packets }} pkts</span>
            <span>{{ time(identity.lastSeen) }}</span>
          </span>
          <span class="observation-groups__actions">
            <UiButton
              v-if="identity.decision !== 'known'"
              size="sm"
              variant="default"
              :disabled="props.pending.has(identity.key)"
              @click="emit('decide', identity.key, 'known')"
            >
              Known
            </UiButton>
            <UiButton
              v-if="identity.decision !== 'ignored'"
              size="sm"
              variant="ghost"
              :disabled="props.pending.has(identity.key)"
              @click="emit('decide', identity.key, 'ignored')"
            >
              Ignore
            </UiButton>
          </span>
        </li>
      </ul>
    </section>

    <p v-if="props.groups.length === 0" class="observation-groups__empty">
      No identities match the current filters.
    </p>
  </div>
</template>

<style scoped>
  .observation-groups {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .observation-groups__heading {
    margin: 0;
    font-size: 0.875rem;
    font-weight: 600;
  }

  .observation-groups__toggle {
    display: flex;
    align-items: center;
    gap: 0.375rem;
    width: 100%;
    background: transparent;
    border: 0;
    border-bottom: 1px solid var(--border-subtle, #2a3441);
    color: inherit;
    font: inherit;
    padding: 0.375rem 0;
    cursor: pointer;
    text-align: left;
  }

  .observation-groups__toggle:focus-visible {
    outline: 2px solid var(--accent-primary, #4a9eff);
    outline-offset: 2px;
  }

  .observation-groups__chevron {
    color: var(--text-muted, #8b9bab);
    width: 1ch;
  }

  .observation-groups__count {
    margin-left: auto;
    color: var(--text-muted, #8b9bab);
    font-weight: 400;
  }

  .observation-groups__list {
    list-style: none;
    margin: 0;
    padding: 0;
  }

  .observation-groups__row {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto auto;
    gap: 0.75rem;
    align-items: center;
    padding: 0.375rem 0;
    border-bottom: 1px solid color-mix(in srgb, var(--border-subtle, #2a3441) 50%, transparent);
    font-size: 0.875rem;
  }

  .observation-groups__title {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .observation-groups__new {
    margin-left: 0.375rem;
    padding: 0 0.3rem;
    border-radius: 999px;
    background: var(--accent-primary, #4a9eff);
    color: #04121f;
    font-size: 0.6875rem;
    font-weight: 700;
    text-transform: uppercase;
  }

  .observation-groups__meta {
    display: flex;
    gap: 0.75rem;
    color: var(--text-muted, #8b9bab);
    font-size: 0.8125rem;
    white-space: nowrap;
  }

  .observation-groups__actions {
    display: flex;
    gap: 0.375rem;
  }

  .observation-groups__empty {
    color: var(--text-muted, #8b9bab);
    font-size: 0.875rem;
  }

  /* Stack the row on narrow screens; the three columns cannot share 390px. */
  @media (max-width: 599px) {
    .observation-groups__row {
      grid-template-columns: 1fr;
      gap: 0.25rem;
    }

    .observation-groups__actions {
      justify-content: flex-start;
    }
  }
</style>
