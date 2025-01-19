<script setup lang="ts">
  import { computed, watch, ref, PropType } from 'vue';
  import TriggerCondition from './TriggerCondition.vue';
  import Selection from '../input/Selection.vue';
  import Dropdown from '../controls/Dropdown.vue';
  import { store } from '../../store/index';
  import { Device } from '@/types/device';
  import {
    AutomationTrigger,
    AutomationTriggerAction,
    AutomationTriggerActions,
    AutomationTriggerCondition,
    AutomationTriggerConditions,
  } from '@/types/automation';
  import { isValid, EditableTriggerCondition, EditableActionTrigger, ActionType } from '../../contracts/automations';
  import { capitalizeText } from '../../modules/formatters/text.formatter';
  import { EventActions, OpenPanelEvent } from '@/types/events.type';
  import { emitCloseLastPanel, emitClosePanel, emitOpenPanel } from '@/mixins/useAutomationsEventBus';
  import ActionViewer from './actions/ActionViewer.vue';
  import ButtonPanel from '@/components/controls/ButtonPanel.vue';
  import { createButtons, createNewActionDropdownItems } from '../../configs/automation/trigger-dropdown.config';

  const props = defineProps({
    id: { type: String },
    trigger: {
      type: Object as PropType<AutomationTrigger>,
      default: {} as AutomationTrigger,
    },
  });

  // TODO: can be refactor to some automation context
  const conditions = ref<AutomationTriggerConditions>({} as AutomationTriggerConditions);
  const actions = ref<AutomationTriggerActions>({} as AutomationTriggerActions);

  const trigger = ref<AutomationTrigger>(props.trigger);

  const dropDownActionItems = computed(() => createNewActionDropdownItems((e: ActionType) => addNewAction(e)));

  const buttonPanelItems = computed(() => {
    return createButtons([
      {
        name: 'Save',
        click: save,
        disabled: isValid(trigger.value) == false,
      },
      {
        name: 'Delete',
        click: remove,
        disabled: isValid(trigger.value) == false,
      },
    ]);
  });

  watch(
    () => props.trigger,
    () => {
      if (
        props.trigger.actions != undefined &&
        props.trigger.actions.every((action) => action.id != '') // refactor
      ) {
        actions.value = JSON.parse(JSON.stringify(props.trigger.actions)) as AutomationTriggerAction[];
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

  function addNewCondition() {
    conditions.value.push(new EditableTriggerCondition());
  }

  function removeTriggerCondition(condition: AutomationTriggerCondition) {
    conditions.value = conditions.value.filter((c) => c != condition);
  }

  const exposesList = computed(() => {
    const device = store.getters['hub/findDevice'](props.id) as Device;
    if (device == null) {
      return [];
    }
    return Object.assign(
      {},
      ...Object.values(device.exposes).map((e) => ({
        [e.name]: e.name,
      }))
    );
  });

  function deleteAction() {
    actions.value = [];
  }

  function SaveAction(currentAction: AutomationTriggerAction, updatedAction: AutomationTriggerAction) {
    const idx = actions.value.findIndex((a) => a == currentAction);
    if (idx != -1) {
      trigger.value.actions[idx] = updatedAction;
    } else {
      trigger.value.actions.push(updatedAction);
    }
  }

  function addNewAction(actionType: ActionType) {
    emitOpenPanel(createActionOpenPanelEvent(new EditableActionTrigger(actionType), true));
  }

  const actionEvents = (currentAction: AutomationTriggerAction): EventActions => {
    return {
      delete: (e) => {
        deleteAction();
      },
      save: (a) => {
        SaveAction(currentAction, a);
      },
    };
  };

  function createActionOpenPanelEvent(action: AutomationTriggerAction, editMode: boolean): OpenPanelEvent {
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
</script>

<template>
  <!-- TODO:  -->
  <!-- if automation for device exists message user else we overwrite it -->

  <div class="row pb-3">
    <div class="col">
      <ButtonPanel :buttons="buttonPanelItems" />
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
    <Fieldset legend="When" :toggleable="true" :collapsed="true">
      <DataTable :value="conditions" selectionMode="single">
        <Column header="Condition">
          <template #body="slotProps">
            <TriggerCondition
              :id="props.id"
              :name="slotProps.data.name"
              :operator="slotProps.data.equality"
              :data="slotProps.data.value"
              @update:name="(newValue) => (slotProps.data.name = newValue)"
              @update:value="(newValue) => (slotProps.data.value = newValue)"
              @update:operator="(newValue) => (slotProps.data.equality = newValue)">
            </TriggerCondition>
          </template>
        </Column>
        <Column class="col-sm-1">
          <template #body="slotProps">
            <Button icon="pi pi-trash" variant="text" rounded @click="removeTriggerCondition(slotProps.data)" />
          </template>
        </Column>
      </DataTable>
      <div class="pt-4 flex align-items-center justify-content-center">
        <Button style="width: 99%" icon="pi pi-plus" label="Add condition" @click="addNewCondition" text size="small" />
      </div>
    </Fieldset>
    <div class="pt-2"></div>
    <Fieldset legend="Then" :toggleable="true" :collapsed="true">
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
