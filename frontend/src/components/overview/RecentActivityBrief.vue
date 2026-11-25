<script setup lang="ts">
  /* Compact three-event preview for Overview; the full filterable feed stays available. */
  import { computed } from 'vue';
  import { useActivityFeed } from '@/composables/useActivityFeed';
  import { findDevice } from '@/composables/useHub';
  import { exposeDisplayLabel, exposeLabel, formatExposeValue } from '@/domain/exposes';
  import { clockTime } from '@/domain/time';
  import UiIcon from '@/components/ui/UiIcon.vue';

  const { enabled, entries } = useActivityFeed();

  const rows = computed(() =>
    entries.value.slice(0, 3).map((entry) => {
      const device = findDevice(entry.deviceId);
      const change = entry.changes[0];
      const expose = change && device?.exposes[change.expose];
      return {
        id: entry.id,
        name: device?.friendly_name ?? entry.deviceId,
        time: clockTime(entry.at),
        label: change ? (expose ? exposeDisplayLabel(expose) : exposeLabel(change.expose)) : 'changed',
        value: change ? (expose ? formatExposeValue(expose, change.to) : String(change.to ?? '—')) : '',
        more: Math.max(0, entry.changes.length - 1),
      };
    })
  );
</script>

<template>
  <div v-if="rows.length" class="nh-brief">
    <div v-for="row in rows" :key="row.id" class="nh-brief-row">
      <time>{{ row.time }}</time>
      <b>{{ row.name }}</b>
      <span>{{ row.label }} <template v-if="row.value">· {{ row.value }}</template><template v-if="row.more"> +{{ row.more }}</template></span>
    </div>
  </div>
  <p v-else-if="!enabled" class="nh-muted nh-brief-empty"><UiIcon name="info" />Activity recording is off. Turn it on in the activity card.</p>
  <p v-else class="nh-muted nh-brief-empty"><UiIcon name="clock" />Waiting for device changes…</p>
</template>

<style scoped>
  .nh-brief {
    display: flex;
    flex-direction: column;
  }
  .nh-brief-row {
    display: flex;
    align-items: baseline;
    gap: 0.55rem;
    padding: 0.3rem 0;
    font-size: 0.8rem;
    color: var(--nh-text-muted);
    min-width: 0;
  }
  .nh-brief-row time {
    width: 3.2rem;
    flex: none;
    font-family: var(--nh-font-mono);
    font-size: 0.7rem;
    color: var(--nh-text-muted);
  }
  .nh-brief-row b {
    color: var(--nh-text);
    font-weight: 600;
    white-space: nowrap;
  }
  .nh-brief-row span {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .nh-brief-link {
    display: inline-flex;
    align-items: center;
    gap: 0.2rem;
    align-self: flex-start;
    margin-top: 0.35rem;
    font-size: 0.75rem;
    font-weight: 600;
    color: var(--nh-accent);
    text-decoration: none;
  }
  .nh-brief-empty {
    display: flex;
    align-items: center;
    gap: 0.35rem;
    font-size: 0.8rem;
  }
</style>
