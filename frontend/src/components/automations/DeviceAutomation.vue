<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import { useRouter } from 'vue-router';
  import InputBox from '../input/InputBox.vue';
  import { store } from '../../store/index';
  import { Device } from '@/types/device';
  import { Automation, AutomationAction, AutomationTrigger } from '@/types/automation.type.js';
  import AutomationStatus from '@/components/automations/schedule/AutomationStatus.vue';
  import ActionButton from '../input/ActionButton.vue';
  import Panel from '../controls/Panel.vue';
  import ButtonPanel from '@/components/controls/ButtonPanel.vue';
  import { createEditAutomationButtonItems } from '../../configs/automation/trigger-dropdown.config';
  import { emitCloseLastPanel } from '@/mixins/useAutomationsEventBus';
  import { emitOpenSchedulerPanelEvent, emitOpenTriggerPanelEvent } from '@/contracts/panel-events';
  import { DataTableRowClickEvent } from 'primevue';
  import { formatTriggerConditions } from '@/transformers/automation/trigger-transformers';
  import { emitOpenConfirmationDialog } from '@/contracts/dialog-events';
  import { useAutomationsLoader } from '@/mixins/composables/useAutomationLoader';
  import {
    canTriggerManually,
    EditableAutomationTrigger,
    DeviceAutomation,
    findActionExposes,
  } from '@/contracts/automations';

  import { triggerAutomation } from '@/services/automation-trigger.service';

  const emit = defineEmits(['cancel']);

  const props = defineProps({
    id: String,
  });

  const router = useRouter();
  const automation = ref<Automation>({} as Automation);
  const isInViewMode = ref<boolean>(true);

  const buttonPanelItems = computed(() => {
    const isActionValid =
      automation.value.triggers?.length == 0 &&
      automation.value.triggers?.filter((k) => k.actions.length != 0).length == automation.value.triggers.length;
    return createEditAutomationButtonItems(
      () => saveAutomation(),
      () => deleteAutomation(),
      () => openScheduler(automation.value),
      isActionValid,
      isActionValid,
      isActionValid
    );
  });

  const automations = useAutomationsLoader();

  // look for changes in props.id or store automations
  watch(
    [() => props.id, automations],
    ([id, updatedAutomations]) => {
      if (!id || updatedAutomations.length === 0) return;

      const found = updatedAutomations.find((a) => a.id === id) as Automation | undefined;
      if (found) {
        // deep clone so editing doesn’t mutate the  source
        automation.value = JSON.parse(JSON.stringify(found)) as DeviceAutomation;
      } else {
        // if automation doesn’t exist yet, create new one
        const device = store.getters['hub/findDevice'](id) as Device | undefined;
        const newAutomation = new DeviceAutomation();
        if (device) {
          newAutomation.id = device.id;
          newAutomation.friendlyname = device.friendly_name;
        }
        automation.value = newAutomation;
        createNewTrigger();
      }
    },
    { immediate: true }
  );

  function createNewTrigger() {
    const newTrigger = EditableAutomationTrigger.create();
    emitOpenTriggerPanelEvent(automation.value.id, newTrigger, saveTrigger, deleteTrigger);
  }

  function openScheduler(automation: Automation) {
    emitOpenSchedulerPanelEvent(automation);
  }

  function cancel() {
    if (isInViewMode.value) {
      emit('cancel');
    } else {
      createCloseLastPanelEvent();
    }
  }

  function rowClicked(event: DataTableRowClickEvent): void {
    const trigger = automation.value.triggers[event.index];

    emitOpenTriggerPanelEvent(automation.value.id, trigger, saveTrigger, deleteTrigger);
  }

  function onComponentDisplayed() {
    isInViewMode.value = false;
  }
  function onComponentHidden() {
    isInViewMode.value = true;
  }
  function createCloseLastPanelEvent() {
    emitCloseLastPanel();
  }

  function saveAutomation() {
    store.dispatch('automations/save', automation.value as Automation);
    router.push('/viewer');
  }

  function deleteAutomation() {
    var sourceAutomation = store.getters['automations/find'](props.id);
    if (sourceAutomation != undefined) {
      store.dispatch('automations/delete', automation.value.id);
      // todo; alert message box to ask user
      router.push('/viewer');
    }
  }

  function deleteTrigger(trigger: AutomationTrigger): void {
    automation.value.triggers = automation.value.triggers.filter((e: AutomationTrigger) => e != trigger);
  }

  function saveTrigger(trigger: AutomationTrigger): void {
    const idx = automation.value.triggers.indexOf(trigger);
    if (idx == -1) {
      automation.value.triggers.push(trigger);
    } else {
      automation.value.triggers[idx] = trigger;
    }
  }

  const deviceNameFromId = (id: string): string => {
    const device = store.getters['hub/findDevice'](id) as Device;
    if (device == undefined) {
      return '';
    }
    return device.friendly_name;
  };

  function getActionDescription(trigger: AutomationTrigger): string {
    if (trigger.actions.every((a: AutomationAction) => a.id == '')) {
      return '<EMPTY>';
    }

    const res = trigger.actions.map(
      (a: AutomationAction) => `${deviceNameFromId(a.id)}.${findActionExposes(a).join('')}`
    );
    return res.join(',');
  }

  function openDeleteTriggerConfirmationDialog(trigger: AutomationTrigger) {
    const props = {
      title: 'Question',
      message: `Delete trigger ${trigger.name} ?`,
    };
    emitOpenConfirmationDialog(() => deleteTrigger(trigger), props);
  }
</script>

<template>
  <div class="relative">
    <!-- TODO: find better way to do this
      we have 2 components that use the same template and toggle from the if isinVieMode -->
    <Button
      icon="pi pi-times"
      size="large"
      variant="text"
      rounded
      class="absolute top-0 right-0 z-5"
      @click="cancel()" />
    <div
      v-bind:style="{
        display: isInViewMode ? 'block' : 'none',
      }">
      <div class="grid">
        <div class="col-12 xl:col-8 lg:col-8 sm:col-8 pb-3">
          <InputBox label="Id" :disabled="true" :value="automation.id" />
        </div>
        <div class="col-12 pb-3">
          <div class="col-12 xl:col-4 lg:col-5 p-0">
            <InputBox label="Friendly Name" :disabled="true" :value="automation.friendlyname" class="w-full" />
          </div>
        </div>
        <div class="col-12 pb-3">
          <div class="col-12 xl:col-4 lg:col-5 p-0">
            <InputBox
              label="Description"
              @updated="(v) => (automation.description = v)"
              :value="automation.description"
              class="w-full" />
          </div>
        </div>
        <div class="col-12 xl:col-8 lg:col-8 sm:col-8 pb-3">
          <AutomationStatus :automation="automation" :clickToOpen="true" />
        </div>
      </div>
      <ButtonPanel :buttons="buttonPanelItems" class="pb-3 pt-3" severity="secondary" />
      <DataTable :value="automation.triggers" @row-click="rowClicked" selectionMode="single">
        <Column field="action" header="Action">
          <template #body="slotProps">
            {{ getActionDescription(slotProps.data) }}
          </template>
        </Column>
        <Column field="conditions" header="Conditions">
          <template #body="slotProps">
            <span v-html="formatTriggerConditions(slotProps.data)" />
            <!-- show action button if we can trigger automation manually -->
            <ActionButton
              v-if="canTriggerManually(slotProps.data)"
              label="Run"
              icon="pi pi-play"
              :action="() => triggerAutomation(automation, slotProps.data.name)" />
          </template>
        </Column>
        <Column style="width: 3rem">
          <template #header="slotProps">
            <Button icon="pi pi-plus" variant="text" rounded @click="createNewTrigger()" />
          </template>
          <template #body="slotProps">
            <Button
              icon="pi pi-trash"
              variant="text"
              rounded
              @click="openDeleteTriggerConfirmationDialog(slotProps.data)" />
          </template>
        </Column>
      </DataTable>
      <div class="pt-4 flex gap-2">
        <Button label="Add Trigger" outlined rounded size="small" @click="createNewTrigger()" />
      </div>
    </div>
    <div
      v-bind:style="{
        display: isInViewMode == false ? 'block' : 'none',
      }">
      <Panel @close="onComponentHidden" @component-displayed="onComponentDisplayed"> </Panel>
    </div>
  </div>
</template>
