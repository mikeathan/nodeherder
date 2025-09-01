<script setup lang="ts">
  import { computed, ref, watch, PropType } from 'vue';
  import {
    TriggerActionExposeBroadcastMode,
    TriggerActionExposeBroadcastModes,
    AutomationTriggerAction,
    AutomationTriggerActionExpose,
    TriggerActionOperation,
  } from '@/types/automation.type.js';
  import ButtonPanel from '@/components/controls/ButtonPanel.vue';
  import {
    createSaveDeleteButtonItems,
    createTriggerActionModesDropdowItems,
    createTriggerActionOperatorsDropdowitems,
  } from '../../../configs/automation/trigger-dropdown.config';
  import DeviceSelector from '@/components/controls/DeviceSelector.vue';
  import ExposeSelector from '@/components/controls/ExposeSelector.vue';
  import InputBox from '@/components/input/InputBox.vue';
  import Dropdown from '@/components/controls/Dropdown.vue';
  import { featureDevicesFilter, featureExposeFilter } from '@/configs/automation/device.config';
  import ExposeDataInput from '@/components/controls/ExposeDataInput.vue';
  import { createTimeIntervalFromMinutes } from '@/contracts/settings';
  import Toggle from '@/components/input/Toggle.vue';
  import { getIconForType } from '@/modules/formatters/icon.formatter';
  import Menu from 'primevue/menu';
  import { ButtonClickEventType, DropDownItemType } from '@/types/controls.type';

  const props = defineProps({
    action: {
      type: Object as PropType<AutomationTriggerAction>,
      default: {} as AutomationTriggerAction,
      required: true,
    },
    automationId: {
      type: String,
      default: '',
      required: false,
    },
  });

  const emit = defineEmits<{
    (e: 'save', action: AutomationTriggerAction): void;
    (e: 'delete', action: AutomationTriggerAction): void;
  }>();

  const action = ref<AutomationTriggerAction>({} as AutomationTriggerAction);

  const operations = ref<TriggerActionOperation[]>([]);

  const splitCommands = ref<boolean>(props.action.splitCommands ?? false);
  watch(
    () => props.action,
    () => {
      action.value = JSON.parse(JSON.stringify(props.action)) as AutomationTriggerAction;
      if (props.action.delay != undefined) {
        operations.value.push('delay');
      }
    },
    { immediate: true }
  );

  const buttonPanelItems = computed(() => {
    const isActionValid = action.value.id && action.value.exposes.filter((e) => e.data && e.name).length > 0;
    return createSaveDeleteButtonItems(
      () => saveAction(),
      () => removeAction(),
      !isActionValid,
      !isActionValid
    );
  });

  function operationDropdownItems(): DropDownItemType[] {
    return createTriggerActionOperatorsDropdowitems((e: TriggerActionOperation) => addOperation(e));
  }

  function broadcastModeDropdownItems(): DropDownItemType[] {
    return createTriggerActionModesDropdowItems((e: any) => setBroadcastMode(e));
  }
  function addOperation(operation: TriggerActionOperation) {
    // TODO: handle more operations when needed
    if (operation != 'delay') {
      console.error('Unknown operation ', operation);
    }

    operations.value.push(operation);
    action.value.delay = createTimeIntervalFromMinutes(1);
  }

  function addNewExpose() {
    const expose: AutomationTriggerActionExpose = {
      name: '',
      data: null,
    };
    action.value.exposes = [...action.value.exposes, expose];
  }

  function removeOperation(operation: TriggerActionOperation) {
    // TODO: handle more operations when needed
    if (operation != 'delay') {
      console.error('Unknown operation ', operation);
    }

    action.value.delay = undefined;
    operations.value = operations.value.filter((e) => e != operation);
  }

  function deviceSelected(id: string, friendlyName: string) {
    action.value.id = id;

    // reset
    action.value.exposes = [];
    operations.value = [];
    action.value.delay = undefined;
  }

  function dataInputChange(expose: AutomationTriggerActionExpose, value: string) {
    expose.data = value;
  }

  function delayInputChange(value: string) {
    action.value.delay!.value = parseInt(value);
  }

  function exposeSelected(expose: AutomationTriggerActionExpose, name: string) {
    expose.name = name;
    expose.data = null;
  }

  function setBroadcastMode(value: boolean) {
    action.value.splitCommands = value;
  }

  function saveAction() {
    emit('save', action.value);
  }
  function removeAction() {
    emit('delete', action.value);
  }

  function removeTriggerExpose(expose: AutomationTriggerActionExpose) {
    action.value.exposes = action.value.exposes.filter((e) => e != expose);
  }

  function operationsAllowed() {
    return action.value.id != '' && operations.value.length == 0 && action.value.exposes.length != 0;
  }

  function hasMultipleExposes() {
    return action.value.exposes.length > 1;
  }
</script>

<style scoped>
  @media (max-width: 640px) {
    .row {
      flex-direction: column;
    }
    .row > .col {
      width: 100% !important;
    }
  }
</style>

<template>
  <div>
    <ButtonPanel :buttons="buttonPanelItems" />

    <div class="pb-3" />
    <h4>Trigger Action</h4>
    <div class="pb-3" />

    <div class="pb-3">
      <DeviceSelector
        label="Device to trigger"
        :id="action.id"
        @updated="deviceSelected"
        :filter="featureDevicesFilter()" />
    </div>

    <div class="flex align-items-center pb-3 gap-1">
      <Button
        class="text-sm"
        icon="pi pi-plus"
        label="Create"
        @click="addNewExpose()"
        size="small"
        severity="secondary"
        :disabled="action.id == ''" />

      <MenuDropdown
        backgroundColor="var(--p-button-secondary-background)"
        :size="16"
        text="Operations"
        :children="operationDropdownItems()"
        :icon="getIconForType('command')"
        :disabled="!operationsAllowed()" />
      <MenuDropdown
        backgroundColor="var(--p-button-secondary-background)"
        :size="16"
        text="Broadcast"
        :children="broadcastModeDropdownItems()"
        :selected="splitCommands"
        :icon="getIconForType('broadcast')"
        :disabled="!hasMultipleExposes()" />
    </div>

    <ReorderableList
      v-model:items="action.exposes"
      item-key="name"
      :show-handle="true"
      @delete-item="removeTriggerExpose">
      <template #default="{ item }">
        <div class="col sm:col-4">
          <ExposeSelector
            :id="action.id"
            label="Expose"
            @updated="(e) => exposeSelected(item, e)"
            :value="item.name"
            :filter="featureExposeFilter()" />
        </div>
        <div class="col sm:col-4">
          <ExposeDataInput
            :show-presets="true"
            :id="action.id"
            :name="item.name"
            label="Set value"
            @updated="(e) => dataInputChange(item, e)"
            :disabled="item.name === ''"
            :value="item.data" />
        </div>
      </template>
    </ReorderableList>

    <div v-for="operation in operations">
      <!-- temporary for now hardcode to delay operation only -->
      <div style="display: flex; align-items: center; gap: 0.5rem">
        <InputBox
          :label="`Delay in ${action[operation]!.unit}`"
          :is-numeric="true"
          :disabled="action.exposes.length == 0"
          @updated="delayInputChange"
          :value="action[operation]!.value" />

        <Button icon="pi pi-trash" text iconOnly="true" @click="removeOperation(operation)" />
      </div>
    </div>
  </div>
</template>
