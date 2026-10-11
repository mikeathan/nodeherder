<script setup lang="ts">
  /*
   * Page-wide notices: lost hub link (NH-06: values may be stale, controls paused) and an
   * open Zigbee permit-join window with a stop button.
   */
  import { store } from '@/store';
  import { useHub } from '@/composables/useHub';
  import { usePermitJoin } from '@/composables/usePermitJoin';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import { connectionText } from './connection';

  const { connection, isConnected } = useHub();
  const permit = usePermitJoin();
  const reconnect = () => store.dispatch('ws/connect');
</script>

<template>
  <div class="nh-banners" aria-live="polite">
    <div v-if="!isConnected" class="nh-banner" :class="connection === 'connecting' ? 'is-warn' : 'is-danger'" role="status">
      <UiIcon :name="connection === 'connecting' ? 'refresh' : 'offline'" />
      <span><b>{{ connectionText(connection) }}.</b> Values may be out of date and device controls are paused until the hub reconnects.</span>
      <UiButton v-if="connection === 'disconnected'" size="sm" @click="reconnect">Reconnect</UiButton>
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
  .nh-banner > span {
    flex: 1 1 14rem;
  }
  .nh-banner.is-warn {
    background: var(--nh-warn-soft);
  }
  .nh-banner.is-danger {
    background: var(--nh-danger-soft);
  }
  .nh-banner.is-info {
    background: var(--nh-accent-soft);
  }
</style>
