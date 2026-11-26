<script setup lang="ts">
  /*
   * Browser-local cleanup only. Server history and log deletion are intentionally absent:
   * they need endpoints, recovery semantics and NH-06 approval, and dead destructive controls
   * are worse than no control.
   */
  import { computed } from 'vue';
  import { store } from '@/store';
  import { useActivityFeed } from '@/composables/useActivityFeed';
  import { useThemeSettings } from '@/composables/useThemeSettings';
  import { confirm } from '@/composables/useConfirm';
  import type { AppNotification } from '@/types/notification.type';
  import type { LogMessageType } from '@/types/console.type';
  import UiCard from '@/components/ui/UiCard.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';

  const activity = useActivityFeed();
  const theme = useThemeSettings();
  const notifications = computed(() => store.getters['notifications/items']() as AppNotification[]);
  const logs = computed(() => store.getters['console/messages']() as LogMessageType[]);

  async function clearNotifications() {
    if (!notifications.value.length) return;
    const ok = await confirm({
      title: 'Clear notifications?',
      message: 'Removes every inbox item for this browser. This cannot be undone.',
      confirmLabel: 'Clear notifications',
      danger: true,
    });
    if (ok) store.dispatch('notifications/clear');
  }

  async function clearActivity() {
    if (!activity.entries.value.length) return;
    const ok = await confirm({
      title: 'Clear recent activity?',
      message: 'Removes the in-memory activity list on this tab. Recording stays on.',
      confirmLabel: 'Clear activity',
      danger: true,
    });
    if (ok) activity.clear();
  }

  async function clearLogs() {
    if (!logs.value.length) return;
    const ok = await confirm({
      title: 'Clear the log buffer?',
      message: 'Removes the live log buffer on this tab. Archived files on the hub are untouched.',
      confirmLabel: 'Clear buffer',
      danger: true,
    });
    if (ok) store.commit('console/clear');
  }

  function resetAppearance() {
    theme.reset();
  }
</script>

<template>
  <section class="nh-section" aria-labelledby="cleanup-heading">
    <header class="nh-section-head">
      <h2 id="cleanup-heading">Data &amp; cleanup</h2>
      <p class="nh-help">Clear data held by this browser. Hub history and log files are not deleted here.</p>
    </header>

    <UiCard title="This browser" icon="settingsStorage" :heading-level="3">
      <div class="cleanup-row">
        <div class="cleanup-tx">
          <b>Notifications</b>
          <small>{{ notifications.length }} kept · read state and digest are cleared too.</small>
        </div>
        <UiButton :disabled="!notifications.length" icon="delete" @click="clearNotifications">Clear</UiButton>
      </div>
      <div class="cleanup-row">
        <div class="cleanup-tx">
          <b>Recent activity</b>
          <small>{{ activity.entries.value.length }} events in memory · recording stays {{ activity.enabled.value ? 'on' : 'off' }}.</small>
        </div>
        <UiButton :disabled="!activity.entries.value.length" icon="delete" @click="clearActivity">Clear</UiButton>
      </div>
      <div class="cleanup-row">
        <div class="cleanup-tx">
          <b>Live log buffer</b>
          <small>{{ logs.length }} messages · archived hub files are untouched.</small>
        </div>
        <UiButton :disabled="!logs.length" icon="delete" @click="clearLogs">Clear</UiButton>
      </div>
      <div class="cleanup-row">
        <div class="cleanup-tx">
          <b>Appearance</b>
          <small>Theme, accent, density, rounding and text size for this browser.</small>
        </div>
        <UiButton icon="refresh" @click="resetAppearance">Reset</UiButton>
      </div>
    </UiCard>

    <p class="nh-alert is-info" role="note">
      <UiIcon name="info" />
      <span>History and archived log deletion stay on the hub roadmap: they need a recovery design and approval before a destructive UI can ship.</span>
    </p>
  </section>
</template>

<style scoped>
  .cleanup-row {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 0.7rem 0;
    border-bottom: var(--nh-border-w) solid var(--nh-border);
  }
  .cleanup-row:last-child {
    border-bottom: 0;
  }
  .cleanup-tx {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
  }
  .cleanup-tx b {
    font-size: 0.85rem;
  }
  .cleanup-tx small {
    color: var(--nh-text-muted);
    font-size: 0.76rem;
  }
  @media (max-width: 560px) {
    .cleanup-row {
      align-items: flex-start;
      flex-direction: column;
      gap: 0.4rem;
    }
  }
</style>
