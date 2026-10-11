<script setup lang="ts">
  /* Hub link indicator in the header (and panel strip). Text hides on narrow screens. */
  import { computed } from 'vue';
  import { store } from '@/store';
  import { useHub } from '@/composables/useHub';
  import { MCPStatusType } from '@/types/settings.type';
  import { connectionText } from './connection';

  const { connection } = useHub();
  const mcp = computed(() => store.getters['hub/mcpStatus']() as MCPStatusType | null);
  const title = computed(() => {
    const parts = [`Hub: ${connectionText(connection.value)}`];
    if (mcp.value?.running) parts.push(`MCP server running, ${mcp.value.connectedClients ?? 0} client(s)`);
    return parts.join(' · ');
  });
</script>

<template>
  <span class="nh-conn" :class="`is-${connection}`" role="status" :title="title">
    <span class="nh-dot" aria-hidden="true" />
    <span class="nh-conn-txt">{{ connectionText(connection) }}</span>
    <span class="sr-only">{{ title }}</span>
  </span>
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
    font-size: 0.8rem;
    white-space: nowrap;
  }
  .nh-dot {
    width: 0.55rem;
    height: 0.55rem;
    border-radius: 50%;
    background: var(--nh-ok);
    box-shadow: 0 0 0 3px var(--nh-ok-soft);
  }
  .is-connecting .nh-dot {
    background: var(--nh-warn);
    box-shadow: 0 0 0 3px var(--nh-warn-soft);
    animation: nh-pulse 1s infinite;
  }
  .is-disconnected .nh-dot {
    background: var(--nh-danger);
    box-shadow: 0 0 0 3px var(--nh-danger-soft);
  }
  @media (max-width: 760px) {
    .nh-conn-txt {
      display: none;
    }
    .nh-conn {
      width: 2rem;
      justify-content: center;
      padding: 0;
    }
  }
</style>
