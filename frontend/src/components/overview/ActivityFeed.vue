<script setup lang="ts">
  /* Recent activity card with an on/off switch (spec 007 FR-12, AC-27). */
  import { computed, ref } from 'vue';
  import { RouteName } from '@/types/router';
  import { useActivityFeed } from '@/composables/useActivityFeed';
  import { findDevice } from '@/composables/useHub';
  import { exposeLabel, formatExposeValue } from '@/domain/exposes';
  import { clockTime } from '@/domain/time';
  import { ActivityChange, filterActivityEntries } from '@/domain/activity';
  import UiCard from '@/components/ui/UiCard.vue';
  import UiToggle from '@/components/ui/UiToggle.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import UiButton from '@/components/ui/UiButton.vue';

  const { enabled, entries, setEnabled, clear } = useActivityFeed();

  const query = ref('');
  const searchInput = ref<HTMLInputElement | null>(null);
  // Presentation only: recording keeps the full bounded feed while searching.
  const matchingEntries = computed(() => filterActivityEntries(entries.value, query.value, (id) => findDevice(id)?.friendly_name));
  function resetFilter() {
    query.value = '';
    searchInput.value?.focus();
  }

  const rows = computed(() =>
    matchingEntries.value.slice(0, 60).map((e) => {
      const device = findDevice(e.deviceId);
      const text = (c: ActivityChange, v: unknown) => {
        const expose = device?.exposes[c.expose];
        return expose ? formatExposeValue(expose, v) : String(v ?? '—');
      };
      return {
        id: e.id,
        deviceId: e.deviceId,
        name: device?.friendly_name ?? e.deviceId,
        time: clockTime(e.at),
        changes: e.changes.map((c) => ({ label: exposeLabel(c.expose), from: text(c, c.from), to: text(c, c.to) })),
      };
    })
  );
</script>

<template>
  <UiCard title="Recent activity" icon="clock" flush class="nh-activity">
    <template #tools>
      <span v-if="enabled && entries.length" class="nh-muted nh-act-count">{{ entries.length }}</span>
      <UiButton v-if="enabled && entries.length" size="sm" variant="ghost" aria-label="Clear all recent activity" @click="clear">Clear</UiButton>
      <label class="nh-feed-switch">
        <span>{{ enabled ? 'On' : 'Off' }}</span>
        <UiToggle :model-value="enabled" label="Record recent activity" @update:model-value="setEnabled" />
      </label>
    </template>
    <div class="nh-activity-filter" role="search" aria-label="Recent activity devices">
      <label class="nh-activity-search">
        <UiIcon name="search" />
        <span class="sr-only">Filter activity by device name</span>
        <input ref="searchInput" v-model="query" class="nh-input" type="search" placeholder="Filter by device name…" @keydown.esc="resetFilter" />
      </label>
      <UiButton v-if="query" icon="close" size="sm" variant="ghost" label="Clear activity filter" @click="resetFilter" />
      <span v-if="enabled && entries.length" class="nh-activity-results" role="status">{{ rows.length }} of {{ matchingEntries.length }} {{ query.trim() ? 'matches' : 'events' }}</span>
    </div>
    <p v-if="!enabled" class="nh-feed-note"><UiIcon name="info" />Activity recording is off. Turn it on to see device changes as they happen.</p>
    <p v-else-if="entries.length && !matchingEntries.length" class="nh-feed-note" role="status"><UiIcon name="search" />No activity matches this device name.</p>
    <p v-else-if="!rows.length" class="nh-feed-note"><UiIcon name="clock" />Waiting for device changes… They appear here while this tab is open.</p>
    <ol v-else class="nh-feed" aria-live="polite" aria-relevant="additions">
      <li v-for="r in rows" :key="r.id">
        <RouterLink :to="{ name: RouteName.DevicePage, params: { id: r.deviceId } }" class="nh-feed-row">
          <time class="nh-feed-time">{{ r.time }}</time>
          <b class="nh-feed-dev">{{ r.name }}</b>
          <span class="nh-feed-chg">
            <span v-for="c in r.changes" :key="c.label">{{ c.label }} <s>{{ c.from }}</s> → <b>{{ c.to }}</b></span>
          </span>
        </RouterLink>
      </li>
    </ol>
  </UiCard>
</template>

<style scoped>
  .nh-activity-filter { display: flex; align-items: center; flex-wrap: wrap; gap: .5rem; padding: .75rem var(--nh-space); border-bottom: var(--nh-border-w) solid var(--nh-border); }
  .nh-activity-search { display: flex; align-items: center; position: relative; flex: 1 1 12rem; min-width: 0; max-width: 26rem; }
  .nh-activity-search > .nh-ic { position: absolute; left: .65rem; color: var(--nh-text-muted); pointer-events: none; }
  .nh-activity-search input { width: 100%; min-width: 0; padding-left: 2.2rem; }
  .nh-activity-results { margin-left: auto; color: var(--nh-text-muted); font-size: .78rem; }
  .nh-act-count {
    font-size: 0.78rem;
  }
  .nh-feed-switch {
    display: inline-flex;
    align-items: center;
    gap: 0.45rem;
    font-size: 0.8rem;
    color: var(--nh-text-muted);
    cursor: pointer;
  }
  .nh-feed-note {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    margin: 0;
    padding: 1.1rem var(--nh-space);
    color: var(--nh-text-muted);
  }
  .nh-feed {
    list-style: none;
    margin: 0;
    padding: 0;
    max-height: 28rem;
    overflow: auto;
  }
  .nh-feed-row {
    display: grid;
    grid-template-columns: auto minmax(7rem, 13rem) 1fr;
    gap: 0.75rem;
    align-items: baseline;
    padding: 0.5rem var(--nh-space);
    text-decoration: none;
    border-bottom: var(--nh-border-w) solid var(--nh-border);
    font-size: 0.85rem;
  }
  .nh-feed-row:hover {
    background: var(--nh-surface-2);
  }
  li:first-child .nh-feed-row {
    animation: nh-flash 1.2s;
  }
  .nh-feed-time {
    font-variant-numeric: tabular-nums;
    color: var(--nh-text-muted);
    font-size: 0.78rem;
  }
  .nh-feed-dev {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .nh-feed-chg {
    display: flex;
    flex-wrap: wrap;
    gap: 0.2rem 0.9rem;
    color: var(--nh-text-muted);
    min-width: 0;
  }
  .nh-feed-chg s {
    opacity: 0.6;
  }
  .nh-feed-chg b {
    color: var(--nh-text);
  }
  @media (max-width: 640px) {
    .nh-feed-row {
      grid-template-columns: auto minmax(0, 1fr);
    }
    .nh-feed-chg {
      grid-column: 1 / -1;
    }
  }
</style>
