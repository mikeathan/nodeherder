<script setup lang="ts">
import { computed, ref, watchEffect, watch, PropType, reactive } from "vue";
import { AutomationTriggerAction } from "@/types/automation";
import { toMillisecs, toMinutes } from "@/modules/formatters/time.formatter";
import ButtonPanel from "@/components/controls/ButtonPanel.vue";
import { createSaveDeleteButtonItems, createTriggerActionOperatorsDropdowitems } from "../../../configs/automation/trigger-dropdown.config";
import DeviceSelector from "@/components/controls/DeviceSelector.vue";
import ExposeSelector from "@/components/controls/ExposeSelector.vue";
import InputBox from '@/components/input/InputBox.vue';
import Dropdown from "@/components/controls/Dropdown.vue";
import {
  featureDevicesFilter,
  featureExposeFilter,
} from "@/configs/automation/device.config";
import ExposeDataInput from "@/components/controls/ExposeDataInput.vue";
import { TriggerActionOperation } from "@/contracts/automations";

const props = defineProps({
  action: {
    type: Object as PropType<AutomationTriggerAction>,
    default: {} as AutomationTriggerAction,
    required: true,
  },
  automationId: {
    type: String,
    default: '',
    required: false
  },
});

const emit = defineEmits<{
  (e: "save", action: AutomationTriggerAction): void;
  (e: "delete", action: AutomationTriggerAction): void;
}>();

const action = reactive({ ...props.action });
const operations = ref<TriggerActionOperation[]>([]);


watch(
  () => props.action,
  () => {

    if (props.action.delay) {
      action.delay = toMinutes(props.action.delay)
      operations.value.push('delay');
    }
  }, { immediate: true }
)

const buttonPanelItems = computed(() => {
  const isActionValid = action.data && action.property && action.id;

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
  operations.value.push(operation)
}

function removeOperation(operation: TriggerActionOperation) {
  action[operation] = null;
  operations.value = operations.value.filter(e => e != operation);
}

function deviceSelected(id: string, friendlyName: string) {
  action.id = id;
  action.friendlyname = friendlyName;

  // reset
  action.property = "";
  action.data = null;
  action.delay = null;
  action.steps = [];
}

function dataInputChange(value: string) {
  action.data = value;
}

function delayInputChange(value: string) {
  action.delay = parseInt(value);
}

function exposeSelected(name: string) {
  action.property = name;

  // reset
  action.data = null;
  action.delay = null;
}


function saveAction() {
  if (action.delay) {
    action.delay = toMillisecs(action.delay)
  }
  emit('save', action)
}
function removeAction() {
  emit('delete', action);
}

</script>

<template>
  <div class="row pb-2">
    <ButtonPanel :buttons="buttonPanelItems">
      <Dropdown :items="dropdownItems" class-name="btn-light" :disabled="action.id == ''">
        Add Operation
      </Dropdown>
    </ButtonPanel>
  </div>

  <h5>Trigger Action</h5>

  <div class="row pb-2">
    <DeviceSelector label="Device to trigger" :id="action.id" @updated="deviceSelected"
      :filter="featureDevicesFilter()">
    </DeviceSelector>
  </div>

  <div class="row pb-2">
    <ExposeSelector :id="action.id" label="Expose" @updated="exposeSelected" :value="action.property"
      :filter="featureExposeFilter()">
    </ExposeSelector>
  </div>

  <div class="row ">
    <div class="col-sm-6">
      <ExposeDataInput :show-presets="true" :id="action.id" :name="action.property" label="Set value"
        @updated="dataInputChange" :disabled="action.property == ''" :value="action.data"></ExposeDataInput>
    </div>
    <div v-for="operation in operations">
      <div class="row">
        <div class="col-sm-4">
          <InputBox label="Delay in minutes" :is-numeric="true" :disabled="action.property == ''"
            @updated="delayInputChange" :value="action[operation]"> </InputBox>
        </div>
        <div class="col-sm-1 pt-4">
          <span class="fa fa-trash-alt fa-sm" @click="removeOperation(operation)"> </span>
        </div>
      </div>
    </div>
  </div>
</template>
