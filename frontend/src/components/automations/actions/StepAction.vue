<script setup lang="ts">
import { computed, PropType, reactive } from "vue";
import {
  AutomationTriggerAction,
  AutomationActionStep,
  NumericOperator,
} from "@/types/automation";
import { store } from "../../../store/index";
import { Device, Devices } from "@/types/device";
import { ExposeTypes } from "@/types/device.type";
import Dropdown from "@/components/controls/Dropdown.vue";
import ButtonPanel from "@/components/controls/ButtonPanel.vue";
import Selection from "@/components/input/Selection.vue";

import {
  createStepActionOperatorsDropdowitems,
  createSaveDeleteButtonItems,
} from "../../../configs/automation/trigger-dropdown.config";
import DeviceSelector from "@/components/controls/DeviceSelector.vue";
import ExposeSelector from "@/components/controls/ExposeSelector.vue";
import {
  featureDevicesFilter,
  exposeFilterByType,
  devicesFilterByActionStep,
} from "@/configs/automation/device.config";
import InputBox from '@/components/input/InputBox.vue';
import { NumericOperators } from "@/contracts/automations";

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

const action = reactive({ ...props.action });
const dropdownItems = computed(() =>
  createStepActionOperatorsDropdowitems((e: NumericOperator) => addStep(e))
);
const buttonPanelItems = computed(() => {
  const actionIsValid =
    action.data && action.property && action.id && action.steps.length != 0;
  return createSaveDeleteButtonItems(
    () => saveAction(),
    () => removeAction(),
    !actionIsValid,
    !actionIsValid
  );
});

const emit = defineEmits<{
  (e: "save", action: AutomationTriggerAction): void;
  (e: "delete", action: AutomationTriggerAction): void;
}>();

function deviceNameFromId(step: AutomationActionStep): string {
  const device = store.getters["devices/find"](step.id) as Device;
  if (device == undefined) {
    return "";
  }
  return device.friendly_name;
}

function stepDeviceSelected(id: string, friendlyName: string, step: AutomationActionStep) {
  step.id = id;
}

function stepPropertySelected(value: string, step: AutomationActionStep) {
  if (action.steps.length == 1) {
    action.property = value;
  }

  step.property = value;
}

function saveAction(): void {
  emit("save", action);
}

function removeAction(): void {
  emit("delete", action);
}

function actionDataChanged(value: any) {
  action.data = value;
}

function addStep(numericOperator: NumericOperator) {
  const newStep: AutomationActionStep = {
    operator: numericOperator,
    property: "",
    id: "",
  };
  action.steps.push(newStep);
}

function removeStep(step: AutomationActionStep) {
  action.steps = action.steps.filter((c) => c != step);
}

function deviceSelected(deviceId: string, friendlyName: string) {
  action.id = deviceId;
  action.friendlyname = friendlyName;
  action.steps = [];
}

</script>

<template>
  <!-- Edit mode -->
  <!-- action controls -->
  <div class="row pb-3">
    <div class="col">
      <ButtonPanel :buttons="buttonPanelItems">
        <Dropdown :items="dropdownItems" class-name="btn-light" :disabled="action.id == ''">
          Add Operation
        </Dropdown>
      </ButtonPanel>
    </div>
  </div>

  <h4>Step Action</h4>
  <div class="pb-3" />


  <!-- device select box  -->
  <div class="row pb-2">
    <DeviceSelector label="Device to trigger" @updated="deviceSelected" :id="action.id"
      :filter="featureDevicesFilter()">
    </DeviceSelector>
  </div>

  <!-- data input box  -->
  <div class="row ">
    <div class="col-sm-3">
      <InputBox label="Set value" :is-numeric="true" :value="action.data" @updated="actionDataChanged">
      </InputBox>
    </div>
  </div>

  <!-- Steps -->
  <div v-if="action.steps.length > 0" class="card">

    <div class="card-header">
      <h5>Operations</h5>
    </div>
    <div class="card-body">
      <ul class="list-group list-group-flush">
        <li class="list-group-item" v-for="step in action.steps">
          <div class="container">
            <div class="row">
              <div class="col-sm-2">
                <Selection :value="step.operator" :items="NumericOperators" @updated="o => step.operator = o">
                </Selection>
              </div>
              <!-- step id  -->
              <div class="col col-xl-4" v-if="step.id == ''">
                <DeviceSelector @updated="(id, name) => stepDeviceSelected(id, name, step)"
                  :filter="devicesFilterByActionStep(props.automationId, action, step)"></DeviceSelector>
              </div>
              <div v-else class="col col-xl-3">
                {{ deviceNameFromId(step) }}
              </div>
              <!-- step property -->
              <div class="col col-xl-6">
                <ExposeSelector :id="step.id" @updated="(v) => stepPropertySelected(v, step)" :value="step.property"
                  :filter="exposeFilterByType(ExposeTypes.Numeric)" position="center"></ExposeSelector>
              </div>
              <div class="col-md-1 col-sm-1">
                <div class="d-grid d-md-auto">
                  <a class="btn btn-sm  btn-light " role="button">
                    <span class="fa fa-trash-alt fa-sm" @click="removeStep(step)"></span>
                  </a>
                </div>
              </div>
            </div>
          </div>
        </li>
      </ul>
    </div>
  </div>
</template>
