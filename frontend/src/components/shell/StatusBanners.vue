<script setup lang="ts">
  /*
   * Content-level notices. Connection and session details live in the header status pill;
   * the snapshot banner is the one place that states how old the visible data is.
   */
  import { computed } from 'vue';
  import { store } from '@/store';
  import { usePermitJoin } from '@/composables/usePermitJoin';
  import { useHub } from '@/composables/useHub';
  import { useNow } from '@/composables/useClock';
  import { startHubConnection } from '@/composables/useHubConnection';
  import { relativeTime } from '@/domain/time';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import UiButton from '@/components/ui/UiButton.vue';

  const permit = usePermitJoin();
  const { connection, initialized } = useHub();
  const now = useNow();
  const syncedAt = computed(() => store.getters['hub/lastSyncedAt']() as number | null);
  const snapshotVisible = computed(
    () => initialized.value && (connection.value === 'reconnecting' || connection.value === 'disconnected') && syncedAt.value !== null
  );
  const snapshotTime = computed(() =>
    syncedAt.value === null ? '' : new Date(syncedAt.value).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  );
  const snapshotAge = computed(() =>
    syncedAt.value === null ? '' : relativeTime(new Date(syncedAt.value).toISOString(), now.value)
  );
</script>

<template>
  <div class="nh-banners" aria-live="polite">
    <div v-if="snapshotVisible" class="nh-banner is-warn" role="status">
      <UiIcon name="offline" />
      <span><b>Live updates paused.</b> Showing the snapshot from {{ snapshotTime }} ({{ snapshotAge }}). Values may have changed.</span>
      <UiButton size="sm" @click="startHubConnection">Retry</UiButton>
    </div>
    <div v-if="permit.active.value" class="nh-banner is-info" role="status">
      <UiIcon name="join" />
      <span><b>Permit join is open</b> for new Zigbee devices · {{ permit.label.value }} left. Put the device in pairing mode now.</span>
      <UiButton size="sm" @click="permit.stop">Stop</UiButton>
    </div>
  </div>
</template>

<style scoped>
  .nh-banner {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 0.6rem;
    padding: 0.55rem var(--nh-space);
    font-size: 0.875rem;
    border-bottom: var(--nh-border-w) solid var(--nh-border);
  }
  .nh-banner > span { flex: 1 1 14rem; }
  .nh-banner.is-info { background: var(--nh-accent-soft); }
  .nh-banner.is-warn { background: var(--nh-warn-soft); color: var(--nh-text); }
</style>
