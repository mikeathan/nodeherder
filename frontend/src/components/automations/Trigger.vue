<script setup lang="ts">
  import { computed, watch, ref, PropType } from 'vue';
  import ConditionEditor from './conditions/ConditionEditor..vue';
  import Selection from '../input/Selection.vue';
  import Dropdown from '../controls/Dropdown.vue';
  import { store } from '../../store/index';
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
  import { isValid, createActionFromType, createConditionFromType } from '../../contracts/automations';
  import { capitalizeText } from '../../modules/formatters/text.formatter';
  import { EventActions, OpenPanelEvent } from '@/types/events.type';
  import { emitClosePanel, emitOpenPanel } from '@/mixins/useAutomationsEventBus';
  import ActionViewer from './actions/ActionViewer.vue';
  import ButtonPanel from '@/components/controls/ButtonPanel.vue';
  import { createButtons, createNewActionDropdownItems } from '../../configs/automation/trigger-dropdown.config';
  import { ExposeCategories } from '@/types/device.type';

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

  const dropDownActionItems = computed(() => createNewActionDropdownItems((e: ActionType) => addNewAction(e)));
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
  <div class="row pb-3">
    <div class="col">
      <ButtonPanel :buttons="buttonPanelItems" severity="secondary" />
    </div>
  </div>
  <div class="row" v-if="trigger.name == ''">
    <Selection
      :value="trigger.name"
      text="Select trigger"
      :disabled="trigger.name != ''"
      size="normal"
      @updated="(v) => (trigger.name = v)"
      :items="exposesList">
    </Selection>
  </div>
  <div class="row" v-else>
    <h4 class="">Trigger for {{ capitalizeText(trigger.name) }}</h4>
    <div class="pb-3" />
    <Fieldset legend="When" :toggleable="true" :collapsed="conditions.length == 0">
      <DataTable
        :value="conditions"
        selectionMode="single"
        :reorderableRows="true"
        @row-reorder="onRowReorder"
        dataKey="name">
        <Column rowReorder style="width: 3rem" />
        <Column header="Condition">
          <template #body="slotProps">
            <ConditionEditor
              :item="slotProps.data"
              :id="props.id"
              @update="(item: AutomationCondition) => updateCondition(slotProps.index, item)"
              @delete="removeTriggerCondition(slotProps.data)" />
          </template>
        </Column>
      </DataTable>
      <div class="pt-4 flex align-items-center justify-content-center">
        <Button
          style="width: 99%"
          icon="pi pi-plus"
          label="Add condition"
          @click="addNewCondition('expose')"
          text
          size="small" />
      </div>
    </Fieldset>
    <div class="pt-2"></div>
    <Fieldset legend="Then" :toggleable="true" :collapsed="actions.length == 0">
      <DataTable :value="actions" selectionMode="single">
        <Column header="Actions">
          <template #body="slotProps">
            <ActionViewer
              :automation-id="props.id"
              :item="slotProps.data"
              :edit-events="actionEvents(slotProps.data)"
              @delete="deleteAction()">
            </ActionViewer>
          </template>
        </Column>
        <Column class="col-sm-1">
          <template #body="slotProps">
            <Button icon="pi pi-trash" variant="text" rounded @click="deleteAction()" />
          </template>
        </Column>
      </DataTable>

      <div class="pt-4 flex align-items-center justify-content-center">
        <Dropdown
          :items="dropDownActionItems"
          :disabled="actions.length != 0"
          text
          label="New Action"
          icon="pi pi-plus"
          size="small" />
      </div>
    </Fieldset>
  </div>
</template>
