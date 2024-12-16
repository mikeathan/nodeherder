<script setup lang="ts">
import { computed } from 'vue';
import { RouterLink, useRouter } from 'vue-router';
import { store } from '../../store/index';
import {
  Automation,
  Automations,
} from '@/types/automation';
import AutomationStatus from './schedule/AutomationStatus.vue';

const router = useRouter();

const automations = computed(() => {
  if (
    !store.getters['automations/initialized']() as Boolean
  ) {
    store.dispatch('ws/emit', { event: 'loadAutomations' });
  }
  return store.getters[
    'automations/listAll'
  ]() as Automations;
});

function onDeleteAutomationClick(id: string): void {
  // emit delete event
  store.dispatch('ws/emit', {
    event: 'deleteAutomation',
    message: {
      id: id,
    },
  });
}

function getIndex(item: Automation): number {
  return automations.value.indexOf(item);
}

const navigateToCreator = () => {
  router.push('/creator');
};
</script>

<template>
  <div class="">
    <div class="">
      <div class="mb-4">Movie Information</div>
      <div class="mb-8">
        Morbi tristique blandit turpis. In viverra ligula id
        nulla hendrerit rutrum.
      </div>
      <ul
        class="list-none p-0 m-0"
        v-for="(automation, index) in automations"
        :key="automation.id">
        <li
          class="col md:flex ">
          <div class="col sm:col-1"># {{ index + 1 }}</div>
          <div class="col sm:col-2">
            <RouterLink :to="`/editor/${automation.id}`">{{
              automation.friendlyname
            }}</RouterLink>
          </div>
          <div class="col sm:col-4">
            {{ automation.description }}
          </div>
          <div class="col sm:col-3">
            <AutomationStatus :automation="automation" />
          </div>
          <div class="flex justify-end">
            <Button
              icon="pi pi-trash"
              variant="text"
              rounded
              @click="
                onDeleteAutomationClick(automation.id)
              " />
          </div>
        </li>
      </ul>
    </div>
  </div>
  <Card>
    <template #title>
      <h2>Automations</h2>
    </template>
    <template #content>
      <DataTable size="small" :value="automations">
        <Column header="#">
          <template #body="slotProps">
            {{ getIndex(slotProps.data) + 1 }}
          </template>
        </Column>
        <Column field="friendlyname" header="Name">
          <template #body="slotProps">
            <RouterLink
              :to="`/editor/${slotProps.data.id}`"
              >{{ slotProps.data.friendlyname }}</RouterLink
            >
          </template>
        </Column>
        <Column field="description" header="Description">
          <template #body="slotProps">
            {{ slotProps.data.description }}
          </template>
        </Column>
        <Column field="enabled" header="Enabled">
          <template #body="slotProps">
            <AutomationStatus
              :automation="slotProps.data" />
          </template>
        </Column>
        <Column>
          <template #body="slotProps">
            <Button
              icon="pi pi-trash"
              variant="text"
              rounded
              @click="
                onDeleteAutomationClick(slotProps.data.id)
              " />
          </template>
        </Column>
      </DataTable>

      <div class="pt-3"></div>
      <div class="col md:col-3 sm:col-6">
        <Button
          style="width: 99%"
          icon="pi pi-plus"
          label="Create automation"
          @click="navigateToCreator"
          size="small" />
      </div>
    </template>
  </Card>
</template>
