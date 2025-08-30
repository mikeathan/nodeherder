<script setup lang="ts">
  import { computed, ref, watch, PropType } from 'vue';
  import {
    AutomationTriggerAction,
    AutomationTriggerActionExpose,
    TriggerActionOperation,
  } from '@/types/automation.type.js';
  import ButtonPanel from '@/components/controls/ButtonPanel.vue';
  import {
    createSaveDeleteButtonItems,
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

  const dropdownItems = computed(() =>
    createTriggerActionOperatorsDropdowitems((e: TriggerActionOperation) => addOperation(e))
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

  const enableSplitCommands = ref(true);
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

  .my-checkbox .p-checkbox-box.p-highlight {
    background-color: #f97316; /* example: warning color */
    border-color: #f97316;
  }

  /* Checkbox box background when unchecked */
  .p-checkbox-box {
    background-color: #f0f0f0;
    border-radius: 0.25rem;
    border: 1px solid #ccc;
    width: 1.5rem;
    height: 1.5rem;
  }

  /* Checked state */
  .p-checkbox-box.p-highlight {
    background-color: #22c55e; /* match 'success' severity */
    border-color: #22c55e;
  }

  /* Checkbox icon color */
  .p-checkbox-icon {
    color: white;
    font-size: 1rem;
  }

  /* Hover effect */
  .p-checkbox:hover .p-checkbox-box {
    border-color: #9ff63b; /* match primary hover color */
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

    <div class="flex align-items-center justify-content-left pb-3 gap-2">
      <Button
        icon="pi pi-plus"
        label="Add Expose"
        @click="addNewExpose()"
        size="small"
        severity="secondary"
        :disabled="action.id == ''" />

      <Dropdown
        :items="dropdownItems"
        icon="pi pi-calculator"
        label="Operations"
        severity="secondary"
        size="small"
        :disabled="!operationsAllowed()" />

      <!-- <Toggle :value="true" :valueOn="true" :valueOff="false" label="Split Commands" /> -->
      <ToggleSwitch v-model="enableSplitCommands">
        <template #handle="{ checked }">
          <i :class="['!text-xs pi', { 'pi-check': checked, 'pi-times': !checked }]" />
        </template>
      </ToggleSwitch>
      <!-- <div class="flex items-center gap-2">
        <Checkbox v-model="enableSplitCommands" binary class="my-checkbox p-mr-2" />
        <span class="text-sm font-medium">Split Commands</span>
      </div> -->
    </div>

    <div class="flex align-items-center justify-content-left pb-3">
      <Dropdown :items="dropdownItems" text label="Operations" size="small" :disabled="!operationsAllowed()" />
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

    <!-- <div class="flex items-center pt-4">

      <div class="flex items-center gap-2">
        <Checkbox inputId="ingredient4" name="pizza" value="Onion" />
        <label for="ingredient4"> Split Commands </label>
      </div>
    </div> -->

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
