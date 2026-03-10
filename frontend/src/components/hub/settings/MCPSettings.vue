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
  <div v-if="mcpStatus" class="mt-4">
    <div class="grid align-items-center gap-y-3">
      <!-- Status -->
      <div class="col-12 md:col-3 text-secondary font-bold">Status</div>
      <div class="col-12 md:col-9">
        <Tag v-if="!mcpEnabled" severity="danger" value="Disabled" icon="pi pi-times" />
        <Tag v-else-if="mcpStatus.running" severity="success" value="Running" icon="pi pi-check" />
        <Tag v-else severity="warning" value="Stopped" icon="pi pi-exclamation-triangle" />
      </div>

      <template v-if="mcpEnabled">
        <!-- Server Name -->
        <div class="col-12 md:col-3 text-secondary font-bold">Server Name</div>
        <div class="col-12 md:col-9">{{ mcpStatus.name }}</div>

        <!-- Version -->
        <div class="col-12 md:col-3 text-secondary font-bold">Version</div>
        <div class="col-12 md:col-9">{{ mcpStatus.version }}</div>

        <!-- Connected Clients -->
        <div class="col-12 md:col-3 text-secondary font-bold">Connected Clients</div>
        <div class="col-12 md:col-9">{{ mcpStatus.connectedClients }}</div>
      </template>

      <!-- Actions -->
      <div class="col-12 mt-3 flex gap-2">
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
