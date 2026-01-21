<script setup lang="ts">
  import { store } from '../../store/index';
  import { computed } from 'vue';

  const connectionStatus = computed(() => {
    return store.getters['ws/getConnectionStatus'] || 'disconnected';
  });
</script>

<template>
  <div class="status-indicator" :class="connectionStatus" :title="`Status: ${connectionStatus}`">
    <div class="dot"></div>
    <div class="ring"></div>
  </div>
</template>

<style scoped>
  .status-indicator {
    position: relative;
    width: 12px;
    height: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background-color: var(--text-color-secondary);
    z-index: 2;
    transition: background-color 0.3s ease;
  }

  .ring {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: 100%;
    height: 100%;
    border-radius: 50%;
    opacity: 0;
    z-index: 1;
  }

  /* Connected State */
  .status-indicator.connected .dot {
    background-color: #4ade80; /* bright green */
    box-shadow: 0 0 4px #4ade80;
  }

  .status-indicator.connected .ring {
    border: 2px solid #4ade80;
    animation: pulse-green 2s infinite;
  }

  /* Disconnected State */
  .status-indicator.disconnected .dot {
    background-color: #f87171; /* red */
    box-shadow: 0 0 2px #f87171;
  }

  .status-indicator.disconnected .ring {
    border: 2px solid #f87171;
    animation: pulse-red 2s infinite; /* Slower pulse for error/disconnected? or just static? */
    animation-duration: 3s;
  }

  /* Connecting State */
  .status-indicator.connecting .dot {
    background-color: #fbbf24; /* amber */
  }

  .status-indicator.connecting .ring {
    border: 2px solid #fbbf24;
    border-top-color: transparent;
    animation: spin 1s linear infinite;
  }

  @keyframes pulse-green {
    0% {
      transform: translate(-50%, -50%) scale(0.8);
      opacity: 0.8;
    }
    70% {
      transform: translate(-50%, -50%) scale(2);
      opacity: 0;
    }
    100% {
      transform: translate(-50%, -50%) scale(0.8);
      opacity: 0;
    }
  }

  @keyframes pulse-red {
    0% {
      transform: translate(-50%, -50%) scale(0.8);
      opacity: 0.5;
    }
    50% {
      transform: translate(-50%, -50%) scale(1.2);
      opacity: 0;
    }
    100% {
      transform: translate(-50%, -50%) scale(0.8);
      opacity: 0;
    }
  }

  @keyframes spin {
    0% {
      transform: translate(-50%, -50%) rotate(0deg);
    }
    100% {
      transform: translate(-50%, -50%) rotate(360deg);
    }
  }
</style>
