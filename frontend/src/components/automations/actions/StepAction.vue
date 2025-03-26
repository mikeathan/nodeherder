<script setup lang="ts">
import { computed, PropType, reactive } from 'vue';
import { AutomationActionStep, NumericOperator, AutomationStepAction } from '@/types/automation.type.js';
import { store } from '../../../store/index';
import { Device } from '@/types/device';
import { ExposeCategories, ExposeTypes } from '@/types/device.type';
import Dropdown from '@/components/controls/Dropdown.vue';
import ButtonPanel from '@/components/controls/ButtonPanel.vue';
import Selection from '@/components/input/Selection.vue';

import {
  createStepActionOperatorsDropdowitems,
  createSaveDeleteButtonItems,
} from '../../../configs/automation/trigger-dropdown.config';
import DeviceSelector from '@/components/controls/DeviceSelector.vue';
import ExposeSelector from '@/components/controls/ExposeSelector.vue';
import {
  featureDevicesFilter,
  exposeFilterByTypeAndCategory,
  devicesFilterByActionStep,
} from '@/configs/automation/device.config';
import InputBox from '@/components/input/InputBox.vue';
import { NumericOperators } from '@/contracts/automations';

const props = defineProps({
  action: {
    type: Object as PropType<AutomationStepAction>,
    default: {} as AutomationStepAction,
    required: true,
  },
  automationId: {
    type: String,
    default: '',
    required: false,
  },
});

const action = reactive({ ...props.action });
const dropdownItems = computed(() => createStepActionOperatorsDropdowitems((e: NumericOperator) => addStep(e)));
const buttonPanelItems = computed(() => {
  const actionIsValid = action.data && action.property && action.id && action.steps.length != 0;
  return createSaveDeleteButtonItems(
    () => saveAction(),
    () => removeAction(),
    !actionIsValid,
    !actionIsValid
  );
});

const emit = defineEmits<{
  (e: 'save', action: AutomationStepAction): void;
  (e: 'delete', action: AutomationStepAction): void;
}>();

function deviceNameFromId(step: AutomationActionStep): string {
  const device = store.getters['hub/findDevice'](step.id) as Device;
  if (device == undefined) {
    return '';
  }
  return device.friendly_name;
}

function stepDeviceSelected(id: string, step: AutomationActionStep) {
  step.id = id;
}

function stepPropertySelected(value: string, step: AutomationActionStep) {
  if (action.steps.length == 1) {
    action.property = value;
  }

  step.property = value;
}

function saveAction(): void {
  emit('save', action);
}

function removeAction(): void {
  emit('delete', action);
}

function actionDataChanged(value: any) {
  action.data = value;
}

function addStep(numericOperator: NumericOperator) {
  const newStep: AutomationActionStep = {
    operator: numericOperator,
    property: '',
    id: '',
  };
  action.steps.push(newStep);
}

function removeStep(step: AutomationActionStep) {
  action.steps = action.steps.filter((c) => c != step);
}

function deviceSelected(deviceId: string) {
  action.id = deviceId;
  action.steps = [];
}
</script>

<template>
  <!-- Edit mode -->
  <!-- action controls -->
  <div class="row pb-3">
    <div class="col">

      <ButtonPanel :buttons="buttonPanelItems" />
    </div>
  </div>

  <h4>Step Action</h4>
  <div class="pb-3" />

  <!-- device select box  -->
  <div class="row pb-2">
    <DeviceSelector label="Device to trigger" @updated="deviceSelected" :id="action.id"
      :filter="featureDevicesFilter()" />
  </div>

  <!-- data input box  -->
  <div class="row">
    <div class="col-sm-3">
      <InputBox label="Set value" :is-numeric="true" :value="action.data" @updated="actionDataChanged" />
    </div>
  </div>
  <div class="flex align-items-center justify-content-left pt-3">
    <Dropdown :items="dropdownItems" :disabled="action.id == ''" text label="Operations" icon="pi pi-plus"
      size="small" />
  </div>
  <!-- Steps -->
  <div v-if="action.steps.length > 0">
    <DataTable :value="action.steps">
      <Column header="Operator">
        <template #body="slotProps">
          <Selection :value="slotProps.data.operator" :items="NumericOperators"
            @updated="(o) => (slotProps.data.operator = o)" />
        </template>
      </Column>

      <Column header="Device">
        <template #body="slotProps">
          <div v-if="slotProps.data.id == ''" style="display: flex; align-items: center; gap: 0.5rem">
            <DeviceSelector @updated="(id, name) => stepDeviceSelected(id, slotProps.data)"
              :filter="devicesFilterByActionStep(props.automationId, action, slotProps.data)"></DeviceSelector>
          </div>
          <div v-else style="display: flex; align-items: center; gap: 0.5rem">
            {{ deviceNameFromId(slotProps.data) }}
          </div>
        </template>
      </Column>

      <Column header="Expose">
        <template #body="slotProps">
          <div style="display: flex; align-items: center; gap: 0.5rem">
            <ExposeSelector :id="slotProps.data.id" @updated="(v) => stepPropertySelected(v, slotProps.data)"
              :value="slotProps.data.property" :filter="exposeFilterByTypeAndCategory(ExposeTypes.Numeric, ExposeCategories.Measurement)" />
            <Button icon="pi pi-trash" text iconOnly="true" @click="removeStep(slotProps.data)" />
          </div>
        </template>
      </Column>
    </DataTable>
  </div>
</template>
