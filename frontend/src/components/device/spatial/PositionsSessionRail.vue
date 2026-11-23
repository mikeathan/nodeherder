<script setup lang="ts">
  import type { SpatialSession } from '@/services/spatial-history.service';
  import {
    RECORDINGS_EMPTY_COPY,
    RECORDINGS_LOADING_COPY,
    railMetaLabel,
    railTimeLabel,
  } from '@/domain/spatial/positions-labels';

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

  const when = (session: SpatialSession): string => railTimeLabel(session, props.spanMs);
  const meta = (session: SpatialSession): string => railMetaLabel(session);
</script>

<template>
  <div class="nh-session-rail">
    <p v-if="loading && !sessions.length" class="nh-muted nh-session-empty" role="status">
      {{ RECORDINGS_LOADING_COPY }}
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
      {{ RECORDINGS_EMPTY_COPY }}
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
    position: relative;
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 0.1rem 0.5rem;
    width: 100%;
    min-height: 2.75rem;
    padding: 0.45rem 0.6rem 0.45rem 0.75rem;
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
  /* The same selection shape as the range timeline: a quiet tint plus a leading edge bar. */
  .nh-session-row.is-selected {
    border-color: var(--nh-accent);
    background: color-mix(in srgb, var(--nh-accent) 22%, transparent);
  }
  .nh-session-row.is-selected::before {
    content: '';
    position: absolute;
    inset-block: 0.35rem;
    inset-inline-start: 0.2rem;
    width: 3px;
    border-radius: 2px;
    background: var(--nh-accent);
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
