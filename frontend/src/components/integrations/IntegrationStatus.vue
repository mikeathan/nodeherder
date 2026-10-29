<script setup lang="ts">
  withDefaults(
    defineProps<{ status: string; checked?: string | null; error?: boolean; warning?: boolean; stacked?: boolean }>(),
    { checked: null, error: false, warning: false, stacked: false }
  );
</script>

<template>
  <div class="integration-status" :class="{ 'is-stacked': stacked }">
    <span class="integration-status-message" :class="{ 'is-error': error, 'is-warning': warning && !error }">{{ status }}</span>
    <span v-if="checked" class="integration-status-checked"><template v-if="!stacked"> · </template>Checked {{ checked }}</span>
  </div>
</template>

<style scoped>
  .integration-status {
    color: var(--nh-text-muted);
    overflow-wrap: anywhere;
  }
  .is-stacked span {
    display: block;
    margin-top: 0.15rem;
  }
  .integration-status-message.is-error {
    color: var(--nh-danger);
  }
  .integration-status-message.is-warning {
    color: var(--nh-warning, var(--nh-text-muted));
  }
</style>
