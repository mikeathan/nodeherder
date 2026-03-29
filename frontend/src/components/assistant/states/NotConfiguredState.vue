<script setup lang="ts">
  import { computed } from 'vue';

  const props = defineProps<{
    reason: 'assistant' | 'mcp' | null;
  }>();

  const title = computed(() => {
    if (props.reason === 'mcp') return 'MCP Server Disconnected';
    return 'Assistant API URL not configured';
  });

  const description = computed(() => {
    if (props.reason === 'mcp') return 'The Model Context Protocol (MCP) server is not running or has no connection. Please configure it in ';
    return 'Please configure the Assistant API URL in ';
  });
</script>

<template>
  <div
    class="absolute top-0 left-0 w-full h-full flex flex-column align-items-center justify-content-center px-4 pb-8 fadein z-1">
    <div class="text-center">
      <i 
        class="text-6xl text-color-secondary mb-4" 
        :class="reason === 'mcp' ? 'pi pi-server' : 'pi pi-cog'"
      ></i>
      <h2 class="text-xl font-normal text-color-secondary m-0">{{ title }}</h2>
      <p class="text-color-secondary mt-2">
        {{ description }}
        <router-link to="/settings" class="text-primary no-underline hover:underline cursor-pointer"
          >Settings</router-link
        >.
      </p>
    </div>
  </div>
</template>
