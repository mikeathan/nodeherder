<script setup lang="ts">
  /*
   * MCP server settings (spec 059 FR-03/FR-05): read-only status and identity plus the
   * process controls. Start/Stop/Restart stay pending until the hub reports a new status;
   * the composable owns that lifecycle and releases it on unmount.
   */
  import { computed, onMounted } from 'vue';
  import { store } from '@/store';
  import { MCP_ACTIONS, useMcpControls, type McpAction } from '@/composables/useMcpControls';
  import UiCard from '@/components/ui/UiCard.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiChip from '@/components/ui/UiChip.vue';
  import UiEmpty from '@/components/ui/UiEmpty.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';

  const { status, enabled, pending, error, refresh, run } = useMcpControls();
  const offline = computed(() => store.getters['ws/getConnectionStatus'] !== 'connected');

  /** Status wording is always paired with a tone so colour alone never carries meaning. */
  const statusChip = computed(() => {
    if (!enabled.value) return { label: 'Disabled', tone: 'muted' as const, icon: 'disabled' as const };
    if (status.value?.running) return { label: 'Running', tone: 'ok' as const, icon: 'check' as const };
    return { label: 'Stopped', tone: 'warn' as const, icon: 'warn' as const };
  });

  const clientLabel = computed(() => {
    const count = status.value?.connectedClients ?? 0;
    return `${count} ${count === 1 ? 'client' : 'clients'}`;
  });

  function trigger(action: McpAction) {
    void run(action);
  }

  onMounted(refresh);
</script>

<template>
  <section class="nh-section" aria-labelledby="mcp-heading">
    <header class="nh-section-head">
      <h2 id="mcp-heading">MCP server</h2>
      <p class="nh-help">Assistant tools connect to this hub through the MCP server.</p>
    </header>

    <p v-if="offline" class="nh-alert is-warn" role="status">
      <UiIcon name="offline" />
      <span>Connect to the hub to control the MCP server.</span>
    </p>

    <UiCard title="Server status" icon="settingsServer" :heading-level="3">
      <template #tools>
        <UiButton size="sm" variant="ghost" icon="refresh" :disabled="offline" @click="refresh">Refresh</UiButton>
      </template>

      <dl v-if="status" class="nh-kv">
        <dt>Status</dt>
        <dd><UiChip :tone="statusChip.tone" :icon="statusChip.icon">{{ statusChip.label }}</UiChip></dd>
        <template v-if="enabled">
          <dt>Server name</dt>
          <dd>{{ status.name || '—' }}</dd>
          <dt>Version</dt>
          <dd>{{ status.version || '—' }}</dd>
          <dt>Connected clients</dt>
          <dd>{{ clientLabel }}</dd>
        </template>
      </dl>
      <UiEmpty v-else icon="info" title="Status unavailable" text="The hub has not reported MCP status yet." />

      <p v-if="error" class="nh-alert is-danger" role="alert"><UiIcon name="error" /><span>{{ error }}</span></p>

      <template #foot>
        <UiButton
          v-if="!status?.running"
          variant="primary"
          icon="run"
          :loading="pending === MCP_ACTIONS.start"
          :disabled="offline || pending !== null"
          @click="trigger(MCP_ACTIONS.start)">
          Start
        </UiButton>
        <UiButton
          v-if="status?.running"
          icon="power"
          :loading="pending === MCP_ACTIONS.stop"
          :disabled="offline || pending !== null"
          @click="trigger(MCP_ACTIONS.stop)">
          Stop
        </UiButton>
        <UiButton
          v-if="status?.running"
          icon="refresh"
          :loading="pending === MCP_ACTIONS.restart"
          :disabled="offline || pending !== null"
          @click="trigger(MCP_ACTIONS.restart)">
          Restart
        </UiButton>
      </template>
    </UiCard>
  </section>
</template>
