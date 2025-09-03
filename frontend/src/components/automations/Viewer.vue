<script setup lang="ts">
  import { computed } from 'vue';
  import { RouterLink, useRouter } from 'vue-router';
  import { store } from '../../store/index';
  import AutomationStatus from './schedule/AutomationStatus.vue';
  import { Automations } from '@/types/automation.type';
  import { emitOpenConfirmationDialog } from '@/contracts/dialog-events';

  const router = useRouter();

  const automations = computed(() => {
    if (!(store.getters['automations/initialized']() as Boolean)) {
      store.dispatch('ws/emit', { event: 'loadAutomations' });
    }
    return store.getters['automations/listAll']() as Automations;
  });

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

<template>
  <Card>
    <template #title>
      <h2>Automations</h2>
    </template>
    <template #content>
      <div class="flex align-items-center pb-3 gap-1">
        <Button
          icon="pi pi-plus"
          class="text-sm"
          label="Create"
          size="small"
          severity="secondary"
          @click="navigateToCreator" />
      </div>

      <ul class="p-0 m-0 list-none">
        <li
          v-for="automation in automations"
          :key="automation.id"
          class="p-3 border-bottom-1 surface-border flex flex-column md:flex-row md:align-items-center md:justify-content-between gap-3">
          <!-- Details -->
          <div class="flex flex-column gap-1 flex-1">
            <RouterLink
              :to="`/editor/${automation.id}`"
              class="font-medium text-lg text-color no-underline hover:underline">
              {{ automation.friendlyname }}
            </RouterLink>
            <span
              class="text-sm text-secondary overflow-hidden text-overflow-ellipsis white-space-nowrap md:white-space-normal">
              {{ automation.description }}
            </span>
          </div>

          <!-- Actions -->
          <div class="flex align-items-center gap-2">
            <AutomationStatus :automation="automation" />
            <Button
              icon="pi pi-trash"
              variant="text"
              rounded
              @click="openDeleteAutomationConfirmationDialog(automation.id)" />
          </div>
        </li>
      </ul>
    </template>
  </Card>
</template>
