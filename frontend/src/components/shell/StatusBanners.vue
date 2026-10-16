<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import { store } from '@/store';
  import { useHub } from '@/composables/useHub';
  import { usePermitJoin } from '@/composables/usePermitJoin';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import { connectionNotices } from './connection';
  import type { ConnectionNotice } from '@/types/connection.type';
  import { useSession } from '@/composables/useSession';
  import { startHubConnection } from '@/composables/useHubConnection';

  const { connection } = useHub();
  const permit = usePermitJoin();
  const { validation } = useSession();
  const notices = computed(() => connectionNotices(connection.value, validation.value, store.getters['hub/loadStatus']));
  const dismissed = ref(false);
  watch(notices, () => { dismissed.value = false; });

  function retry(notice: ConnectionNotice) {
    if (notice.kind === 'transport') store.dispatch('ws/connect');
    else void startHubConnection();
  }
  function dismiss() {
    dismissed.value = true;
    document.querySelector<HTMLElement>('.nh-conn')?.focus();
  }
</script>

<template>
  <aside v-if="notices.length && !dismissed" class="nh-connection-notices" aria-label="Connection notices" aria-live="polite">
    <div v-for="notice in notices" :key="notice.kind" class="nh-banner" :class="notice.kind === 'transport' ? 'is-danger' : 'is-warn'" role="status">
      <UiIcon name="offline" />
      <span><b>{{ notice.title }}</b> {{ notice.text }}</span>
      <UiButton size="sm" @click="retry(notice)">{{ notice.retry }}</UiButton>
    </div>
    <UiButton size="sm" variant="ghost" class="nh-notice-dismiss" @click="dismiss">Dismiss connection notices</UiButton>
  </aside>
  <div class="nh-banners" aria-live="polite">
    <div v-if="permit.active.value" class="nh-banner is-info" role="status">
      <UiIcon name="join" />
      <span><b>Permit join is open</b> for new Zigbee devices · {{ permit.label.value }} left. Put the device in pairing mode now.</span>
      <UiButton size="sm" @click="permit.stop">Stop</UiButton>
    </div>
  </div>
</template>

<style scoped>
  .nh-connection-notices {
    position: fixed;
    z-index: 45;
    right: max(1rem, env(safe-area-inset-right));
    bottom: max(1rem, env(safe-area-inset-bottom));
    width: min(34rem, calc(100vw - 2rem));
    max-height: calc(100dvh - var(--nh-header-h) - 2rem);
    overflow: auto;
    border: var(--nh-border-w) solid var(--nh-border);
    border-radius: var(--nh-radius-m);
    background: var(--nh-surface);
    box-shadow: var(--nh-shadow-2);
  }
  .nh-notice-dismiss { margin: 0.4rem; }
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
  .nh-banner.is-warn { background: var(--nh-warn-soft); }
  .nh-banner.is-danger { background: var(--nh-danger-soft); }
  .nh-banner.is-info { background: var(--nh-accent-soft); }
</style>
