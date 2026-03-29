<script setup lang="ts">
  import { store } from '../../store/index';
  import { computed } from 'vue';

  const connectionStatus = computed(() => store.getters['ws/getConnectionStatus'] || 'disconnected');
  const mcpStatus = computed(() => store.getters['hub/mcpStatus']());
  
  const isMcpRunning = computed(() => mcpStatus.value?.running);
  const isMcpConnected = computed(() => isMcpRunning.value && mcpStatus.value.connectedClients > 0);

  const tooltip = computed(() => {
    let text = `Hub: ${connectionStatus.value}`;
    if (isMcpRunning.value) {
      const clientCount = mcpStatus.value.connectedClients || 0;
      text += ` | MCP: ${isMcpConnected.value ? 'Active' : 'Idle'} (${clientCount} clients)`;
    }
    return text;
  });
</script>

<template>
  <div 
    class="status-dot" 
    :class="[
      connectionStatus, 
      { 'mcp-connected': isMcpConnected, 'mcp-idle': isMcpRunning && !isMcpConnected }
    ]" 
    :title="tooltip"
  ></div>
</template>

<style scoped>
  .status-dot {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    transition: all 0.3s ease;
    background-color: #64748b; /* Neutral Gray */
    position: relative;
    flex-shrink: 0;
    /* Ensure space for the halo */
    margin: 4px; 
  }

  /* Core WS Connection States */
  .status-dot.connected { background-color: #22c55e; }    /* Green */
  .status-dot.disconnected { background-color: #ef4444; } /* Red */
  .status-dot.connecting { background-color: #eab308; animation: blink 0.8s infinite alternate; }

  /* STATE 1: MCP Connected & Active (Solid Blue Halo) */
  .status-dot.mcp-connected {
    box-shadow: 
      0 0 0 2px #121212,   /* Match your header BG color */
      0 0 0 3.5px #3b82f6; /* Solid Blue */
  }

  /* STATE 2: MCP Running but Idle/Searching (Pulsing Gray/Amber Halo) */
  .status-dot.mcp-idle {
    box-shadow: 
      0 0 0 2px #121212, 
      0 0 0 3.5px rgba(203, 213, 225, 0.4); /* Faint Gray */
    animation: mcp-searching 2s infinite ease-in-out;
  }

  @keyframes mcp-searching {
    0%, 100% { box-shadow: 0 0 0 2px #121212, 0 0 0 3px rgba(203, 213, 225, 0.3); }
    50% { box-shadow: 0 0 0 2px #121212, 0 0 0 4.5px rgba(234, 179, 8, 0.5); } /* Pulse to Amber */
  }

  @keyframes blink {
    from { opacity: 1; }
    to { opacity: 0.5; }
  }
</style>