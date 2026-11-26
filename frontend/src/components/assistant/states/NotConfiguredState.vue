<script setup lang="ts">
  /* Configuration gate with a direct route to the settings section that fixes it. */
  import { computed } from 'vue';
  import { RouteName } from '@/types/router';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';

  const props = defineProps<{ reason: 'assistant' | 'mcp' | null }>();

  const title = computed(() => (props.reason === 'mcp' ? 'MCP server is not connected' : 'Assistant address is not set'));
  const description = computed(() =>
    props.reason === 'mcp'
      ? 'Start the MCP server and connect a client before asking the assistant about your home.'
      : 'Set the assistant service address in Settings before starting a conversation.'
  );
</script>

<template>
  <div class="nh-not-config">
    <span class="nh-not-config-ic"><UiIcon :name="reason === 'mcp' ? 'settings' : 'assistant'" /></span>
    <h2>{{ title }}</h2>
    <p>{{ description }}</p>
    <UiButton :to="{ name: RouteName.Settings }" icon="settings">Open settings</UiButton>
  </div>
</template>

<style scoped>
  .nh-not-config {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 0.55rem;
    padding: 2rem var(--nh-space);
    text-align: center;
  }
  .nh-not-config-ic {
    width: 3rem;
    height: 3rem;
    border-radius: 1rem;
    display: grid;
    place-items: center;
    background: var(--nh-surface-2);
    color: var(--nh-text-muted);
  }
  .nh-not-config-ic .nh-ic {
    width: 1.6rem;
    height: 1.6rem;
  }
  .nh-not-config h2 {
    font-size: 1.1rem;
  }
  .nh-not-config p {
    max-width: 30rem;
    color: var(--nh-text-muted);
    font-size: 0.85rem;
  }
</style>
