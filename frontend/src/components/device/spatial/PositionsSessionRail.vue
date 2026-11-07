<script setup lang="ts">
  import { computed } from 'vue';
  import type { SpatialSession } from '@/services/spatial-history.service';
  import { clock, durationLabel } from '@/domain/spatial/playback';

  const props = defineProps<{
    /** Newest first; the caller owns ordering. */
    sessions: readonly SpatialSession[];
    selectedId: string;
    /** Requested but not yet loaded; shown as pending, never as confirmed. */
    pendingId?: string;
    /** Preset span, so long ranges show the date on each row. */
    spanMs: number;
    loading?: boolean;
  }>();
  const emit = defineEmits<{ select: [id: string] }>();

  const showDate = computed(() => props.spanMs > 86_400_000);

  function when(session: SpatialSession) {
    const options = { date: showDate.value, seconds: false };
    return `${clock(session.from, options)} – ${clock(session.to, options)}`;
  }

  function meta(session: SpatialSession) {
    const slots = session.slots.length ? session.slots.join(', ') : 'none';
    return `${durationLabel(session.to - session.from)} · ${session.points} points · slot ${slots}`;
  }
</script>

<template>
  <div class="nh-session-rail">
    <p v-if="loading && !sessions.length" class="nh-muted nh-session-empty" role="status">
      Loading recordings…
    </p>
    <ol v-else-if="sessions.length" class="nh-session-list" aria-label="Recorded sessions">
      <li v-for="session in sessions" :key="session.id">
        <button
          type="button"
          class="nh-session-row"
          :class="{
            'is-selected': session.id === selectedId,
            'is-pending': session.id === pendingId && session.id !== selectedId,
          }"
          :aria-pressed="session.id === selectedId"
          :aria-busy="session.id === pendingId || undefined"
          :data-session-id="session.id"
          @click="emit('select', session.id)">
          <span class="nh-session-when">{{ when(session) }}</span>
          <span v-if="session.capped" class="nh-chip is-warn nh-session-capped">capped</span>
          <span class="nh-session-meta">{{ meta(session) }}</span>
        </button>
      </li>
    </ol>
    <p v-else class="nh-muted nh-session-empty">
      No recorded positions in this range. Enable “Record history” on the device and let a target move.
    </p>
  </div>
</template>

<style scoped>
  .nh-session-list {
    display: grid;
    gap: 0.35rem;
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .nh-session-row {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 0.1rem 0.5rem;
    width: 100%;
    min-height: 2.75rem;
    padding: 0.45rem 0.6rem;
    border: 1px solid var(--nh-border);
    border-radius: var(--nh-radius-m);
    background: var(--nh-surface);
    color: inherit;
    text-align: start;
    cursor: pointer;
  }
  .nh-session-row:hover {
    border-color: var(--nh-border-strong);
  }
  .nh-session-row.is-selected {
    border-color: var(--nh-accent);
    background: var(--nh-accent-soft);
  }
  .nh-session-row.is-pending {
    border-style: dashed;
    opacity: 0.7;
  }
  .nh-session-row:focus-visible {
    outline: 2px solid var(--nh-focus);
    outline-offset: 2px;
  }
  .nh-session-when {
    font-weight: 600;
    font-variant-numeric: tabular-nums;
  }
  .nh-session-meta {
    grid-column: 1 / -1;
    color: var(--nh-text-muted);
    font-size: 0.78rem;
  }
  .nh-session-capped {
    justify-self: end;
  }
  .nh-session-empty {
    margin: 0;
    font-size: 0.85rem;
  }
  @media (max-width: 899px) {
    .nh-session-row {
      min-height: 44px;
    }
  }
</style>
