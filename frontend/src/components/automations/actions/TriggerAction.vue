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

  const operationDropdownItems = computed(() =>
    createTriggerActionOperatorsDropdowitems((e: TriggerActionOperation) => addOperation(e))
  );

  const broadcastModeDropdownItems = computed(() =>
    createTriggerActionModesDropdowItems(
      (e: TriggerActionExposeBroadcastMode) => setBroadcastMode(e),
      action.value.splitCommands ?? false
    )
  );
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
    {{ action }}
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
        icon="pi pi-plus"
        label="Create"
        @click="addNewExpose()"
        size="small"
        severity="secondary"
        :disabled="action.id == ''" />

      <Dropdown
        :items="operationDropdownItems"
        icon="pi pi-calculator"
        label="Operations"
        severity="secondary"
        size="small"
        :disabled="!operationsAllowed()" />

      <Dropdown
        :items="broadcastModeDropdownItems"
        :selected="action.splitCommands"
        icon="pi pi-send"
        severity="secondary"
        label="Broadcast"
        size="small" />
    </div>

    <!-- <div class="flex items-center pt-4">
      <Toggle
        :value="splitCommands"
        @update="updateSplitCommands"
        :valueOn="true"
        :valueOff="false"
        left-label="Batch"
        right-label="Individual" />
    </div> -->

    <div class="flex align-items-center justify-content-left pb-3">
      <Dropdown :items="operationDropdownItems" text label="Operations" size="small" :disabled="!operationsAllowed()" />
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
