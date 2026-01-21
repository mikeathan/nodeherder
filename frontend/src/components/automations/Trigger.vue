<script setup lang="ts">
  import { computed, watch, ref, PropType } from 'vue';
  import ConditionEditor from './conditions/ConditionEditor..vue';
  import Selection from '../input/Selection.vue';
  import Dropdown from '../controls/Dropdown.vue';
  import { store } from '@/store';
  import { Device } from '@/types/device';
  import {
    AutomationTrigger,
    AutomationActions,
    AutomationCondition,
    AutomationTriggerConditions,
    AutomationAction,
    ActionType,
    ConditionType,
  } from '@/types/automation.type.js';
  import { getActionTypes, getConditionTypes } from '../../contracts/automations';
  import { isValid, createActionFromType, createConditionFromType } from '../../contracts/automations';
  import { capitalizeText } from '../../modules/formatters/text.formatter';
  import { EventActions, OpenPanelEvent } from '@/types/events.type';
  import { emitClosePanel, emitOpenPanel } from '@/mixins/useAutomationsEventBus';
  import ActionViewer from './actions/ActionViewer.vue';
  import ButtonPanel from '@/components/controls/ButtonPanel.vue';
  import { createButtons } from '../../configs/automation/trigger-dropdown.config';
  import { ExposeCategories } from '@/types/device.type';
  import { emitOpenConfirmationDialog } from '@/contracts/dialog-events';

  const props = defineProps({
    id: { type: String },
    trigger: {
      type: Object as PropType<AutomationTrigger>,
      default: {} as AutomationTrigger,
    },
  });

  // TODO: can be refactor to some automation context
  const conditions = ref<AutomationTriggerConditions>({} as AutomationTriggerConditions);
  const actions = ref<AutomationActions>({} as AutomationActions);
  const trigger = ref<AutomationTrigger>(props.trigger);

  const buttonPanelItems = computed(() => {
    return createButtons([
      {
        label: 'Save',
        command: save,
        disabled: isValid(trigger.value) == false,
      },
      {
        label: 'Delete',
        command: remove,
        disabled: isValid(trigger.value) == false,
      },
    ]);
  });

  function openDeleteConditionConfirmationDialog(condition: AutomationCondition) {
    const props = {
      title: 'Question',
      message: `Delete condition ${condition.type} ?`,
    };
    emitOpenConfirmationDialog(() => removeTriggerCondition(condition), props);
  }

  function openDeleteActionConfirmationDialog() {
    const props = {
      title: 'Question',
      message: `Delete action ?`,
    };
    emitOpenConfirmationDialog(() => deleteAction(), props);
  }

  watch(
    () => props.trigger,
    () => {
      if (
        props.trigger.actions != undefined &&
        props.trigger.actions.every((action: AutomationAction) => action.id != '') // refactor
      ) {
        actions.value = JSON.parse(JSON.stringify(props.trigger.actions)) as AutomationAction[];
      }
      conditions.value = JSON.parse(JSON.stringify(props.trigger.conditions)) as AutomationTriggerConditions;
    },
    { immediate: true }
  );

  const emit = defineEmits<{
    (e: 'save', trigger: AutomationTrigger): void;
    (e: 'delete', trigger: AutomationTrigger): void;
  }>();

  function save() {
    trigger.value.conditions = conditions.value;
    trigger.value.actions = actions.value;
    emit('save', trigger.value);
    emitClosePanel('Trigger');
  }

  function remove() {
    emit('delete', trigger.value);
    emitClosePanel('Trigger');
  }

  function addNewCondition(type: ConditionType) {
    conditions.value.push(createConditionFromType(type));
  }

  function removeTriggerCondition(condition: AutomationCondition) {
    conditions.value = conditions.value.filter((c: AutomationCondition) => c != condition);
  }

  function updateCondition(idx: number, newCondition: AutomationCondition) {
    conditions.value[idx] = newCondition;
  }

  const exposesList = computed(() => {
    const device = store.getters['hub/findDevice'](props.id) as Device;
    if (device == null) {
      return [];
    }
    return Object.assign(
      {},
      ...Object.values(device.exposes)
        .filter((ex) => ex.category != ExposeCategories.Config)
        .map((e) => ({
          [e.name]: e.name,
        }))
    );
  });

  function deleteAction() {
    actions.value = [];
  }

  function SaveAction(currentAction: AutomationAction, updatedAction: AutomationAction) {
    const idx = actions.value.findIndex((a: AutomationAction) => a == currentAction);
    if (idx != -1) {
      trigger.value.actions[idx] = updatedAction;
    } else {
      trigger.value.actions.push(updatedAction);
    }
  }

  function addNewAction(actionType: ActionType) {
    emitOpenPanel(createActionOpenPanelEvent(createActionFromType(actionType), true));
  }

  const actionEvents = (currentAction: AutomationAction): EventActions => {
    return {
      delete: (e) => {
        deleteAction();
      },
      save: (a) => {
        SaveAction(currentAction, a);
      },
    };
  };

  function createActionOpenPanelEvent(action: AutomationAction, editMode: boolean): OpenPanelEvent {
    return {
      name: 'ActionEditor',
      args: {
        automationId: props.id,
        item: action,
        editMode: editMode,
      },
      events: actionEvents(action),
    };
  }

  function onRowReorder(e: { value: AutomationCondition[] }) {
    conditions.value = e.value;
  }
</script>

<template>
  <!-- TODO:  -->
  <!-- if automation for device exists message user else we overwrite it -->
  <div class="grid pb-3">
    <div class="col-12">
      <ButtonPanel :buttons="buttonPanelItems" severity="secondary" />
    </div>
  </div>
  <div class="grid" v-if="trigger.name == ''">
    <Selection
      :value="trigger.name"
      text="Select trigger"
      :disabled="trigger.name != ''"
      size="normal"
      @updated="(v) => (trigger.name = v)"
      :items="exposesList">
    </Selection>
  </div>
  <div class="grid" v-else>
    <div class="col-12">
      <h4>Trigger for {{ capitalizeText(trigger.name) }}</h4>
      <div class="pt-2" />

    <!-- Conditions -->

    <!-- Conditions -->

    <Fieldset legend="Conditions" :toggleable="true" :collapsed="conditions.length == 0">
      <DataTable
        :value="conditions"
        selectionMode="single"
        :reorderableRows="true"
        @row-reorder="onRowReorder"
        dataKey="name">
        <Column rowReorder style="width: 3rem" />
        <Column>
          <template #body="slotProps">
            <ConditionEditor
              :item="slotProps.data"
              :id="props.id"
              @update="(item: AutomationCondition) => updateCondition(slotProps.index, item)"
              @delete="removeTriggerCondition(slotProps.data)" />
          </template>
        </Column>
        <Column style="width: 1rem; text-align: center">
          <template #body="slotProps">
            <Button
              icon="pi pi-trash"
              variant="text"
              rounded
              small
              @click="openDeleteConditionConfirmationDialog(slotProps.data)" />
          </template>
        </Column>
      </DataTable>
      <div class="pt-4 flex gap-2">
        <div v-for="type in getConditionTypes()" :key="type">
          <Button :label="'Add ' + type" outlined rounded size="small" @click="addNewCondition(type)" />
        </div>
      </div>
    </Fieldset>
    <div class="pt-2" />

    <!-- Actions -->

    <Fieldset legend="Actions" :toggleable="true" :collapsed="actions.length == 0">
      <DataTable :value="actions" selectionMode="single">
        <Column>
          <template #body="slotProps">
            <ActionViewer
              :automation-id="props.id"
              :item="slotProps.data"
              :edit-events="actionEvents(slotProps.data)"
              @delete="openDeleteActionConfirmationDialog()">
            </ActionViewer>
          </template>
        </Column>
        <Column style="width: 3rem">
          <template #body="slotProps">
            <Button icon="pi pi-trash" variant="text" rounded @click="openDeleteActionConfirmationDialog()" />
          </template>
        </Column>
      </DataTable>
      <div v-if="actions.length == 0" class="pt-4 flex gap-2">
        <div v-for="type in getActionTypes()" :key="type">
          <Button :label="'Add ' + type" outlined rounded size="small" @click="addNewAction(type)" />
        </div>
      </div>
    </Fieldset>
    </div>
  </div>
</template>
