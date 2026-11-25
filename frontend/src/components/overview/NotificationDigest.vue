<script setup lang="ts">
  /*
   * Session digest for Overview: at most three unread events since the user last dismissed
   * it. Dismissal is per newest item, so a genuinely new event brings the card back without
   * replaying toasts.
   */
  import { computed, ref } from 'vue';
  import { store } from '@/store';
  import type { AppNotification } from '@/types/notification.type';
  import { safeStorage } from '@/utils/storage';
  import { useNotificationPanel } from '@/composables/useNotificationPanel';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import RelativeTime from '@/components/ui/RelativeTime.vue';

  const DISMISSED_KEY = 'nodeherder_digest_dismissed';
  const { show } = useNotificationPanel();
  const dismissed = ref(safeStorage.getItem(DISMISSED_KEY) ?? '');
  const items = computed(() => store.getters['notifications/unread'](3) as AppNotification[]);
  const visible = computed(() => items.value.length > 0 && items.value[0].id !== dismissed.value);

  function dismiss() {
    const newest = items.value[0];
    if (!newest) return;
    dismissed.value = newest.id;
    safeStorage.setItem(DISMISSED_KEY, newest.id);
  }
</script>

<template>
  <section v-if="visible" class="nh-digest" aria-labelledby="nh-digest-title">
    <div class="nh-digest-head">
      <UiIcon name="bell" />
      <h2 id="nh-digest-title">While you were away</h2>
      <span>{{ items.length }} unread</span>
      <div class="nh-digest-actions">
        <UiButton size="sm" variant="ghost" @click="dismiss">Dismiss</UiButton>
        <UiButton size="sm" @click="show">Review all</UiButton>
      </div>
    </div>
    <ul>
      <li v-for="item in items" :key="item.id">
        <UiIcon :name="item.severity === 'error' ? 'error' : item.severity === 'warn' ? 'warn' : item.severity === 'success' ? 'check' : 'info'" />
        <span class="nh-digest-title">{{ item.title }}</span>
        <RelativeTime :value="item.at" />
      </li>
    </ul>
  </section>
</template>

<style scoped>
  .nh-digest {
    border: var(--nh-border-w) solid color-mix(in srgb, var(--nh-accent) 38%, transparent);
    background: linear-gradient(180deg, var(--nh-accent-soft), transparent);
    border-radius: var(--nh-radius-l);
    padding: 0.8rem 0.9rem;
    margin-bottom: var(--nh-space);
  }
  .nh-digest-head {
    display: flex;
    align-items: center;
    gap: 0.45rem;
    flex-wrap: wrap;
  }
  .nh-digest-head h2 {
    font-size: 0.9rem;
    font-weight: 700;
  }
  .nh-digest-head > span {
    color: var(--nh-text-muted);
    font-size: 0.72rem;
  }
  .nh-digest-actions {
    margin-left: auto;
    display: flex;
    gap: 0.35rem;
  }
  .nh-digest ul {
    list-style: none;
    margin: 0.5rem 0 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
  }
  .nh-digest li {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    font-size: 0.8rem;
    color: var(--nh-text-muted);
    min-width: 0;
  }
  .nh-digest li > .nh-ic {
    width: 0.9rem;
    height: 0.9rem;
    flex: none;
  }
  .nh-digest-title {
    flex: 1;
    min-width: 0;
    color: var(--nh-text);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>
