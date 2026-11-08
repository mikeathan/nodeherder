<script setup lang="ts">
  /* Content-level notices. Connection and session details live in the header status pill. */
  import { usePermitJoin } from '@/composables/usePermitJoin';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import UiButton from '@/components/ui/UiButton.vue';

  const permit = usePermitJoin();
</script>

<template>
  <div class="nh-banners" aria-live="polite">
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
</style>
