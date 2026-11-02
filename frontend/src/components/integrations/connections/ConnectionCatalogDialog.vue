<script setup lang="ts">
  // Discovery is a dialog boundary; catalog grouping/search lives in its own component.
  import type { RouteLocationRaw } from 'vue-router';
  import type { ConnectionCatalog, ConnectionInstance } from '@/types/integrations';
  import UiDialog from '@/components/ui/UiDialog.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import IntegrationCatalog from '../catalog/IntegrationCatalog.vue';

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
    <IntegrationCatalog
      :catalog="catalog"
      :hidden="hidden"
      :busy="busy"
      :provider-route="providerRoute"
      @add="$emit('add', $event)"
      @restore="$emit('restore', $event)" />
    <template #footer>
      <UiButton :disabled="busy || !catalog?.available" icon="refresh" @click="$emit('reload')"
        >Reload providers</UiButton
      >
    </template>
  </UiDialog>
</template>
