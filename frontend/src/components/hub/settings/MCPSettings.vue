<script setup lang="ts">
  import { ConnectionStatus } from '@/types/connection.type';
  import { computed, watch } from 'vue';
  import { store } from '../../../store/index';
  import { MCPStatusType } from '@/types/settings.type';
  import ActionButton from '@/components/input/ActionButton.vue';

  const mcpStatus = computed(() => {
    return store.getters['hub/mcpStatus']() as MCPStatusType | null;
  });

  const mcpEnabled = computed(() => {
    return store.getters['hub/mcpConfig']() as boolean;
  });

  const connectionStatus = computed(() => {
    return store.getters['ws/getConnectionStatus'];
  });

  watch(
    connectionStatus,
    (newValue) => {
      if (newValue === ConnectionStatus.connected) {
        store.dispatch('hub/loadMCPStatus');
      }
    },
    { immediate: true }
  );

  async function restart() {
    store.dispatch('hub/restartMCP');
    await new Promise((resolve) => setTimeout(resolve, 1000));
  }

  async function stop() {
    store.dispatch('hub/stopMCP');
    await new Promise((resolve) => setTimeout(resolve, 1000));
  }

  async function start() {
    store.dispatch('hub/startMCP');
    await new Promise((resolve) => setTimeout(resolve, 1000));
  }
</script>

<template>
  <h3>MCP Server</h3>
  <div class="pt-3" />

  <div v-if="mcpStatus">
    <div class="grid grid-nogutter pb-2">
      <dl class="col-12 md:col-3 text-secondary">
        <dt><strong>Status</strong></dt>
      </dl>
      <div class="col-12 md:col-3">
        <Tag v-if="!mcpEnabled" severity="danger" value="Disabled" />
        <Tag v-else-if="mcpStatus.running" severity="success" value="Running" />
        <Tag v-else severity="warning" value="Stopped" />
      </div>
    </div>

    <template v-if="mcpEnabled">
      <div class="grid grid-nogutter pb-2">
        <dl class="col-12 md:col-3 text-secondary">
          <dt><strong>Server Name</strong></dt>
        </dl>
        <div class="col-12 md:col-3">{{ mcpStatus.name }}</div>
      </div>

      <div class="grid grid-nogutter pb-2">
        <dl class="col-12 md:col-3 text-secondary">
          <dt><strong>Version</strong></dt>
        </dl>
        <div class="col-12 md:col-3">{{ mcpStatus.version }}</div>
      </div>

      <div class="grid grid-nogutter pb-2">
        <dl class="col-12 md:col-3 text-secondary">
          <dt><strong>Connected Clients</strong></dt>
        </dl>
        <div class="col-12 md:col-3">{{ mcpStatus.connectedClients }}</div>
      </div>
    </template>

    <div class="grid grid-nogutter pt-2">
      <div class="col-12 md:col-6 flex gap-2">
        <ActionButton
          v-if="!mcpEnabled || !mcpStatus.running"
          label="Start"
          icon="pi pi-play"
          severity="success"
          :action="start" />
        <ActionButton v-if="mcpEnabled" label="Stop" icon="pi pi-stop" severity="danger" :action="stop" />
        <ActionButton
          v-if="mcpStatus.running"
          label="Restart"
          icon="pi pi-refresh"
          severity="warning"
          :action="restart" />
      </div>
    </div>
  </div>
  <div v-else class="text-secondary">Loading MCP status...</div>
</template>
