<script setup lang="ts">
  /*
   * Durable notification inbox. The bell owns the unread badge and the dialog; Overview's
   * digest opens the same dialog through `useNotificationPanel`, so there is one inbox.
   */
  import { computed } from 'vue';
  import { store } from '@/store';
  import type { AppNotification, NotificationSeverity } from '@/types/notification.type';
  import { useNotificationPanel } from '@/composables/useNotificationPanel';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiDialog from '@/components/ui/UiDialog.vue';
  import UiEmpty from '@/components/ui/UiEmpty.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import RelativeTime from '@/components/ui/RelativeTime.vue';

  const { open, show, hide } = useNotificationPanel();
  const items = computed(() => store.getters['notifications/items']() as AppNotification[]);
  const unread = computed(() => store.getters['notifications/unreadCount']() as number);
  const label = computed(() => (unread.value ? `Notifications, ${unread.value} unread` : 'Notifications'));

  const icons: Record<NotificationSeverity, 'error' | 'warn' | 'check' | 'info'> = {
    error: 'error',
    warn: 'warn',
    success: 'check',
    info: 'info',
  };

  function markRead(item: AppNotification) {
    if (!item.read) store.dispatch('notifications/markRead', item.id);
  }
  function markAll() {
    store.dispatch('notifications/markAllRead');
  }
  function openAction() {
    hide();
  }
</script>

<template>
  <span class="nh-bell-wrap">
    <UiButton variant="ghost" icon="bell" :label="label" @click="show" />
    <span v-if="unread" class="nh-bell-badge" aria-hidden="true">{{ unread > 9 ? '9+' : unread }}</span>
  </span>
  <span class="sr-only" role="status" aria-live="polite">{{ unread }} notification{{ unread === 1 ? '' : 's' }} need attention</span>

  <UiDialog :open="open" title="Notifications" wide fullscreen-mobile @close="hide">
    <div class="nh-notif-toolbar">
      <p>{{ unread ? `${unread} unread` : 'All caught up' }}<span v-if="items.length"> · {{ items.length }} kept for 7 days</span></p>
      <UiButton v-if="unread" size="sm" variant="ghost" @click="markAll">Mark all read</UiButton>
    </div>

    <ul v-if="items.length" class="nh-notif-list">
      <li v-for="item in items" :key="item.id" :class="{ 'is-unread': !item.read }" class="nh-notif-item">
        <span class="nh-notif-ic" :class="`is-${item.severity}`"><UiIcon :name="icons[item.severity]" /></span>
        <div class="nh-notif-tx">
          <b>{{ item.title }}</b>
          <p v-if="item.body">{{ item.body }}</p>
          <small>
            <RelativeTime :value="item.at" />
            <template v-if="item.at"> · </template>{{ item.category }}
            <template v-if="unread && !item.read"> · unread</template>
          </small>
          <RouterLink v-if="item.action" :to="item.action.to" class="nh-notif-link" @click="openAction">
            {{ item.action.label }}
          </RouterLink>
        </div>
        <UiButton
          v-if="!item.read"
          size="sm"
          variant="ghost"
          :label="`Mark ${item.title} read`"
          @click="markRead(item)">
          Mark read
        </UiButton>
      </li>
    </ul>
    <UiEmpty v-else icon="bell" title="No notifications yet" text="Connection problems, failed operations and other actionable events appear here." />
    <template #footer>
      <span class="nh-notif-note">Kept in this browser. Nothing is sent to the hub.</span>
      <UiButton @click="hide">Done</UiButton>
    </template>
  </UiDialog>
</template>

<style scoped>
  .nh-bell-wrap {
    position: relative;
    display: inline-flex;
  }
  .nh-bell-badge {
    position: absolute;
    top: 0.1rem;
    right: 0.1rem;
    min-width: 1rem;
    height: 1rem;
    padding: 0 0.25rem;
    border-radius: 999px;
    background: var(--nh-danger);
    color: #fff;
    font: 700 0.62rem/1rem var(--nh-font-ui);
    text-align: center;
    border: 2px solid var(--nh-header-bg, var(--nh-bg));
    pointer-events: none;
  }
  .nh-notif-toolbar {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-bottom: 0.6rem;
  }
  .nh-notif-toolbar p {
    flex: 1;
    margin: 0;
    color: var(--nh-text-muted);
    font-size: 0.8rem;
  }
  .nh-notif-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
  }
  .nh-notif-item {
    display: flex;
    gap: 0.65rem;
    align-items: flex-start;
    padding: 0.65rem;
    border-radius: var(--nh-radius-m);
    border: var(--nh-border-w) solid transparent;
  }
  .nh-notif-item.is-unread {
    background: var(--nh-surface-2);
    border-color: var(--nh-border);
  }
  .nh-notif-ic {
    width: 1.8rem;
    height: 1.8rem;
    border-radius: 0.55rem;
    display: grid;
    place-items: center;
    flex: none;
    background: var(--nh-surface-2);
    color: var(--nh-text-muted);
  }
  .nh-notif-ic.is-error {
    background: var(--nh-danger-soft);
    color: var(--nh-danger);
  }
  .nh-notif-ic.is-warn {
    background: var(--nh-warn-soft);
    color: var(--nh-warn);
  }
  .nh-notif-ic.is-success {
    background: var(--nh-ok-soft);
    color: var(--nh-ok);
  }
  .nh-notif-ic.is-info {
    background: var(--nh-info-soft);
    color: var(--nh-info);
  }
  .nh-notif-tx {
    flex: 1;
    min-width: 0;
  }
  .nh-notif-tx b {
    display: block;
    font-size: 0.86rem;
  }
  .nh-notif-tx p {
    margin: 0.1rem 0 0;
    color: var(--nh-text-muted);
    font-size: 0.8rem;
  }
  .nh-notif-tx small {
    display: block;
    margin-top: 0.2rem;
    color: var(--nh-text-muted);
    font-size: 0.7rem;
  }
  .nh-notif-link {
    display: inline-block;
    margin-top: 0.3rem;
    font-size: 0.75rem;
    font-weight: 600;
    color: var(--nh-accent);
  }
  .nh-notif-note {
    margin-right: auto;
    color: var(--nh-text-muted);
    font-size: 0.72rem;
  }
</style>
