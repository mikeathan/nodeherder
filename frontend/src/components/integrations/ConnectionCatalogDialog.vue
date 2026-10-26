<script setup lang="ts">
  // Discovery/configuration is deliberately separate from the everyday results board.
  import type { RouteLocationRaw } from 'vue-router';
  import type { ConnectionCatalog, ConnectionInstance } from '@/types/integrations.type';
  import UiDialog from '@/components/ui/UiDialog.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  defineProps<{
    open: boolean;
    catalog: ConnectionCatalog | null;
    hidden: ConnectionInstance[];
    busy: boolean;
    providerRoute: (key: string) => RouteLocationRaw;
  }>();
  defineEmits<{ close: []; add: [key: string]; restore: [id: string]; reload: [] }>();
</script>
<template>
  <UiDialog :open="open" title="Add integration" wide :dismissible="!busy" @close="$emit('close')">
    <div class="connection-catalog">
      <section v-if="hidden.length" aria-label="Hidden integrations">
        <h2>Hidden integrations</h2>
        <div v-for="item in hidden" :key="item.id" class="connection-catalog-row">
          <div>
            <h3>{{ item.name }}</h3>
            <p>Configured and hidden from this page</p>
          </div>
          <UiButton :disabled="busy" @click="$emit('restore', item.id)">Show on page</UiButton>
        </div>
      </section>
      <section v-for="type in catalog?.types" :key="type.id" :aria-label="type.name">
        <h2>{{ type.name }}</h2>
        <p>{{ type.description }}</p>
        <div
          v-for="provider in catalog?.providers.filter((item) => item.integration_type === type.id)"
          :key="provider.key"
          class="connection-catalog-row">
          <span class="connection-catalog-icon"><UiIcon name="connections" /></span>
          <div>
            <h3>{{ provider.name }}</h3>
            <p v-if="provider.disk_error" role="status">{{ provider.disk_error }}</p>
          </div>
          <div class="connection-catalog-actions">
            <UiButton variant="primary" :disabled="busy || !provider.available" @click="$emit('add', provider.key)"
              >Set up</UiButton
            >
            <UiButton :to="providerRoute(provider.key)" :disabled="busy">Configure operations</UiButton>
          </div>
        </div>
      </section>
      <p v-if="catalog?.available && !catalog.providers.length">
        No providers are available. Check your provider configuration files, then reload.
      </p>
    </div>
    <template #footer
      ><UiButton :disabled="busy || !catalog?.available" icon="refresh" @click="$emit('reload')"
        >Reload providers</UiButton
      ></template
    >
  </UiDialog>
</template>
<style scoped>
  .connection-catalog {
    display: grid;
    gap: 1.5rem;
    min-width: 0;
  }
  .connection-catalog h2 {
    margin: 0 0 0.75rem;
    font-size: 1rem;
  }
  .connection-catalog h3 {
    margin: 0;
    font-size: 0.95rem;
    overflow-wrap: anywhere;
  }
  .connection-catalog p {
    color: var(--nh-text-muted);
    font-size: 0.85rem;
    overflow-wrap: anywhere;
  }
  .connection-catalog-row {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 0.75rem;
    padding: 0.85rem 0;
    border-top: var(--nh-border-w) solid var(--nh-border);
  }
  .connection-catalog-row > div:first-of-type {
    flex: 1;
    min-width: 7rem;
  }
  .connection-catalog-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }
  .connection-catalog-icon {
    display: inline-flex;
    padding: 0.7rem;
    border-radius: 50%;
    background: var(--nh-surface-2);
    color: var(--nh-accent);
  }
</style>
