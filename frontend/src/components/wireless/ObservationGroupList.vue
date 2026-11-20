<script setup lang="ts">
  /*
   * Grouped list view for wireless observations (spec 066).
   *
   * Sections are behaviour classes in a fixed order, not vendors sorted by size. Class is the
   * question the operator is asking ("what is mine, what is outside"), and a fixed order means a
   * changing count cannot reshuffle the headings. Rows use the same component as the radar side
   * panel so a device looks identical in both surfaces.
   */
  import { ref } from 'vue';
  import type { ObservationDecision, ObservationIdentityView } from '@/types/observations.type';
  import type { ObservationClassGroup } from '@/domain/observations/classes';
  import ObservationRow from './ObservationRow.vue';

  const props = defineProps<{
    groups: readonly {
      group: string;
      groupId: ObservationClassGroup;
      identities: ObservationIdentityView[];
    }[];
    /** Identity keys with a save in flight, so their controls report progress. */
    pending: ReadonlySet<string>;
    /** Currently isolated identity, so the row can be marked. */
    selectedKey?: string | null;
    /** Current time in milliseconds, shared by every row's relative label. */
    now: number;
    /** Source id to display name, so rows can say which receiver heard them. */
    sourceNames: ReadonlyMap<string, string>;
    /** Explanation shown when filters leave nothing to render. */
    emptyMessage: string;
  }>();

  const emit = defineEmits<{
    (event: 'decide', key: string, decision: ObservationDecision): void;
    (event: 'select', key: string): void;
  }>();

  /** Sections the operator has folded. Collapsed state is display-only. */
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
</script>

<template>
  <div class="observation-groups">
    <section v-for="group in props.groups" :key="group.groupId" class="observation-groups__section">
      <h2 class="observation-groups__heading">
        <button
          type="button"
          class="observation-groups__toggle"
          :aria-expanded="!isCollapsed(group.groupId)"
          @click="toggleGroup(group.groupId)"
        >
          <span class="observation-groups__chevron" aria-hidden="true">{{ isCollapsed(group.groupId) ? '▸' : '▾' }}</span>
          {{ group.group }}
          <span class="observation-groups__count">{{ group.identities.length }}</span>
        </button>
      </h2>

      <ul v-show="!isCollapsed(group.groupId)" class="observation-groups__list">
        <ObservationRow
          v-for="identity in group.identities"
          :key="identity.key"
          :identity="identity"
          :pending="props.pending.has(identity.key)"
          :selected="props.selectedKey === identity.key"
          :now="props.now"
          :source-names="props.sourceNames"
          @decide="(key, decision) => emit('decide', key, decision)"
          @select="(key) => emit('select', key)"
        />
      </ul>
    </section>

    <p v-if="props.groups.length === 0" class="observation-groups__empty">
      {{ props.emptyMessage }}
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
    border-bottom: 1px solid var(--nh-border);
    color: inherit;
    font: inherit;
    padding: 0.375rem 0;
    cursor: pointer;
    text-align: left;
  }

  .observation-groups__toggle:focus-visible {
    outline: 2px solid var(--nh-accent);
    outline-offset: 2px;
  }

  .observation-groups__chevron {
    color: var(--nh-text-muted);
    width: 1ch;
  }

  .observation-groups__count {
    margin-left: auto;
    color: var(--nh-text-muted);
    font-weight: 400;
    font-variant-numeric: tabular-nums;
  }

  .observation-groups__list {
    list-style: none;
    margin: 0;
    padding: 0;
  }

  .observation-groups__empty {
    color: var(--nh-text-muted);
    font-size: 0.875rem;
  }
</style>
