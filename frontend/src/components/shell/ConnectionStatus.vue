<script setup lang="ts">
  /*
   * Hub link indicator in the header. The pill is the single connection-status surface:
   * a warning badge appears when there is an actionable notice, and clicking opens a small
   * dialog with the details and retry actions. No persistent banners.
   */
  import { computed, ref } from 'vue';
  import { store } from '@/store';
  import { useHub } from '@/composables/useHub';
  import { useSession } from '@/composables/useSession';
  import { startHubConnection } from '@/composables/useHubConnection';
  import type { MCPStatusType } from '@/types/settings.type';
  import type { ConnectionNotice } from '@/types/connection.type';
  import { connectionNotices, connectionText } from './connection';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiDialog from '@/components/ui/UiDialog.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';

  const { connection } = useHub();
  const { validation } = useSession();
  const mcp = computed(() => store.getters['hub/mcpStatus']() as MCPStatusType | null);
  const notices = computed(() =>
    connectionNotices(connection.value, validation.value, store.getters['hub/loadStatus'])
  );
  const open = ref(false);
  const label = computed(() => {
    const parts = [`Hub: ${connectionText(connection.value)}`];
    if (notices.value.length) {
      parts.push(`${notices.value.length} connection ${notices.value.length === 1 ? 'notice' : 'notices'}`);
    }
    if (mcp.value?.running) {
      parts.push(`MCP server running, ${mcp.value.connectedClients ?? 0} client(s)`);
    }
    return parts.join(' · ');
  });
  const liveLabel = computed(() => {
    const issue = notices.value.length
      ? ` · ${notices.value.length} connection ${notices.value.length === 1 ? 'notice' : 'notices'}`
      : '';
    return `${connectionText(connection.value)}${issue}`;
  });

  function retry(notice: ConnectionNotice) {
    if (notice.kind === 'transport') store.dispatch('ws/connect');
    else void startHubConnection();
  }
</script>

<template>
  <button
    type="button"
    class="nh-conn"
    :class="[`is-${connection}`, { 'is-attention': notices.length }]"
    :title="label"
    :aria-label="label"
    aria-haspopup="dialog"
    :aria-expanded="open"
    @click="open = true">
    <span class="nh-dot" aria-hidden="true" />
    <span class="nh-conn-txt">{{ connectionText(connection) }}</span>
    <UiIcon v-if="notices.length" name="warn" class="nh-conn-alert" />
  </button>
  <span class="sr-only" role="status" aria-live="polite">{{ liveLabel }}</span>

  <UiDialog :open="open" title="Connection" @close="open = false">
    <div class="nh-conn-summary">
      <span class="nh-conn" :class="`is-${connection}`" aria-hidden="true">
        <span class="nh-dot" />
        <span class="nh-conn-txt">{{ connectionText(connection) }}</span>
      </span>
      <span v-if="mcp?.running" class="nh-muted">
        MCP server running · {{ mcp.connectedClients ?? 0 }} client(s)
      </span>
    </div>
    <div v-for="notice in notices" :key="notice.kind" class="nh-conn-notice">
      <p><strong>{{ notice.title }}</strong> {{ notice.text }}</p>
      <UiButton size="sm" @click="retry(notice)">{{ notice.retry }}</UiButton>
    </div>
    <p v-if="!notices.length" class="nh-muted">The hub connection is healthy.</p>
    <template #footer>
      <UiButton variant="ghost" @click="open = false">Close</UiButton>
    </template>
  </UiDialog>
</template>

<style scoped>
  .nh-conn {
    display: inline-flex;
    align-items: center;
    gap: 0.45rem;
    height: 2rem;
    padding: 0 0.7rem;
    border-radius: 99px;
    border: var(--nh-border-w) solid var(--nh-border);
    background: var(--nh-surface);
    color: inherit;
    font: inherit;
    font-size: 0.8rem;
    white-space: nowrap;
    min-width: 8.5rem;
    justify-content: center;
    cursor: pointer;
  }
  .nh-conn:hover {
    border-color: var(--nh-border-strong);
  }
  .nh-dot {
    width: 0.55rem;
    height: 0.55rem;
    border-radius: 50%;
    background: var(--nh-text-muted);
  }
  .nh-conn:focus-visible { outline: 2px solid var(--nh-accent); outline-offset: 2px; }
  .is-connected .nh-dot { background: var(--nh-ok); box-shadow: 0 0 0 3px var(--nh-ok-soft); }
  .is-reconnecting .nh-dot {
    background: var(--nh-warn);
    box-shadow: 0 0 0 3px var(--nh-warn-soft);
    animation: nh-pulse 1s infinite;
  }
  .is-disconnected .nh-dot {
    background: var(--nh-danger);
    box-shadow: 0 0 0 3px var(--nh-danger-soft);
  }
  .is-attention .nh-dot {
    background: var(--nh-warn);
    box-shadow: 0 0 0 3px var(--nh-warn-soft);
  }
  .nh-conn-alert {
    width: 0.95rem;
    height: 0.95rem;
    color: var(--nh-warn);
  }
  .is-disconnected .nh-conn-alert {
    color: var(--nh-danger);
  }
  .nh-conn-summary {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.5rem 0.75rem;
  }
  .nh-conn-notice {
    display: grid;
    gap: 0.5rem;
    justify-items: start;
    padding: 0.65rem 0.75rem;
    border: 1px solid var(--nh-border);
    border-radius: var(--nh-radius-m);
    background: var(--nh-surface-2);
  }
  .nh-conn-notice p {
    margin: 0;
  }
  @media (max-width: 760px) {
    .nh-conn-txt {
      display: none;
    }
    .nh-conn {
      width: 2.75rem;
      min-width: 2.75rem;
      height: 2.75rem;
      justify-content: center;
      padding: 0;
    }
  }
</style>
