<script setup lang="ts">
import {
  computed,
  ref,
  watch,
  PropType,
  reactive,
} from 'vue';
import { AutomationTriggerAction } from '@/types/automation';
import {
  toMillisecs,
  toMinutes,
} from '@/modules/formatters/time.formatter';
import ButtonPanel from '@/components/controls/ButtonPanel.vue';
import {
  createSaveDeleteButtonItems,
  createTriggerActionOperatorsDropdowitems,
} from '../../../configs/automation/trigger-dropdown.config';
import DeviceSelector from '@/components/controls/DeviceSelector.vue';
import ExposeSelector from '@/components/controls/ExposeSelector.vue';
import InputBox from '@/components/input/InputBox.vue';
import Dropdown from '@/components/controls/Dropdown.vue';
import {
  featureDevicesFilter,
  featureExposeFilter,
} from '@/configs/automation/device.config';
import ExposeDataInput from '@/components/controls/ExposeDataInput.vue';
import { TriggerActionOperation } from '@/contracts/automations';

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

const action = reactive({ ...props.action });
const operations = ref<TriggerActionOperation[]>([]);

watch(
  () => props.action,
  () => {
    if (props.action.delay) {
      action.delay = toMinutes(props.action.delay);
      operations.value.push('delay');
    }
  },
  { immediate: true }
);

const buttonPanelItems = computed(() => {
  const isActionValid =
    action.data && action.property && action.id;

  return createSaveDeleteButtonItems(
    () => saveAction(),
    () => removeAction(),
    !isActionValid,
    !isActionValid
  );
});

const dropdownItems = computed(() =>
  createTriggerActionOperatorsDropdowitems(
    (e: TriggerActionOperation) => addOperation(e)
  )
);

function addOperation(operation: TriggerActionOperation) {
  operations.value.push(operation);
}

function removeOperation(
  operation: TriggerActionOperation
) {
  action[operation] = null;
  operations.value = operations.value.filter(
    (e) => e != operation
  );
}

function deviceSelected(
  id: string,
  friendlyName: string
) {
  action.id = id;
  action.friendlyname = friendlyName;

  // reset
  action.property = '';
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
    action.delay = toMillisecs(action.delay);
  }
  emit('save', action);
}
function removeAction() {
  emit('delete', action);
}
</script>

<template>
  <ButtonPanel :buttons="buttonPanelItems" />

  <div class="pb-3" />
  <h4>Trigger Action</h4>
  <div class="pb-3" />

  <div class="pb-3">
    <DeviceSelector label="Device to trigger" :id="action.id" @updated="deviceSelected"
      :filter="featureDevicesFilter()" />
  </div>
  <div class="pb-3">
    <ExposeSelector :id="action.id" label="Expose" @updated="exposeSelected" :value="action.property"
      :filter="featureExposeFilter()" />
  </div>
  <div class="flex align-items-center justify-content-left pb-3">
    <Dropdown :items="dropdownItems" :disabled="action.id == ''" text label="Operations" icon="pi pi-plus"
      size="small" />
  </div>
  <div class="row">
    <div class="col sm:col-4">
      <ExposeDataInput :show-presets="true" :id="action.id" :name="action.property" label="Set value"
        @updated="dataInputChange" :disabled="action.property == ''" :value="action.data"></ExposeDataInput>
    </div>
  </div>
  <div v-for="operation in operations">
    <div style="
        display: flex;
        align-items: center;
        gap: 0.5rem;
      ">
      <InputBox label=" Delay in minutes" :is-numeric="true" :disabled="action.property == ''"
        @updated="delayInputChange" :value="action[operation]" />
      <Button icon="pi pi-trash" text iconOnly="true" @click="removeOperation(operation)" />
    </div>
  </div>
</template>
