<script setup lang="ts">
  import { computed, watch } from 'vue';
  import { useRouter } from 'vue-router';
  import { store } from '../../store/index';
  import AutomationStatus from './schedule/AutomationStatus.vue';
  import { emitOpenConfirmationDialog } from '@/contracts/dialog-events';

  const router = useRouter();
  const automations = computed(() => store.getters['automations/listAll']());

  watch(
    () => store.getters['ws/getConnectionStatus'],
    (status) => {
      if (status === 'connected' && !store.getters['automations/initialized']()) {
        store.dispatch('ws/emit', { event: 'loadAutomations' });
      }
    },
    { immediate: true }
  );

  function openDeleteAutomationConfirmationDialog(id: string) {
    const props = {
      title: 'Question',
      message: `Delete automation ${id} ?`,
    };
    emitOpenConfirmationDialog(() => onDeleteAutomationClick(id), props);
  }

  function onDeleteAutomationClick(id: string): void {
    store.dispatch('ws/emit', {
      event: 'deleteAutomation',
      message: { id },
    });
  }

  const navigateToCreator = () => {
    router.push('/creator');
  };
</script>
<style scoped>
  .hover-row:hover,
  .hover-row:focus,
  .hover-row:active {
    background: var(--p-content-hover-background);
  }
</style>
<template>
  <Card>
    <template #title>
      <h2>Automations</h2>
    </template>
    <template #content>
      <div class="flex align-items-center pb-3 gap-1">
        <Button icon="pi pi-plus" label="Create" size="small" severity="secondary" @click="navigateToCreator" />
      </div>
      <DataView :value="automations" layout="list" data-key="id">
        <template #list="slotProps">
          <div
            v-for="automation in slotProps.items"
            :key="automation.id"
            class="flex flex-column md:flex-row md:align-items-center md:justify-content-between p-3 border-bottom-1 surface-border mb-2 cursor-pointer hover-row"
            @click="router.push(`/editor/${automation.id}`)">
            <!-- Details -->
            <div class="flex flex-column gap-1 flex-1">
              <span class="font-medium text-lg text-primary">{{ automation.friendlyname }}</span>
              <span class="text-sm text-secondary">{{ automation.description }}</span>
            </div>
            <!-- Actions -->
            <div class="flex align-items-center gap-2" @click.stop>
              <AutomationStatus :automation="automation" />
              <Button
                icon="pi pi-trash"
                variant="text"
                rounded
                @click="openDeleteAutomationConfirmationDialog(automation.id)" />
            </div>
          </div>
        </template>
      </DataView>
    </template>
  </Card>
</template>
