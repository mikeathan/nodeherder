<script setup lang="ts">
import {
  computed,
  ref,
  watch,
  PropType,
  reactive,
} from 'vue';
import { AutomationTriggerAction, AutomationTriggerActionExpose } from '@/types/automation';
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
      // action.delay = toMinutes(props.action.delay);
      operations.value.push('delay');
    }
  },
  { immediate: true }
);

const buttonPanelItems = computed(() => {
  const isActionValid = false;
  // action.exposes.filter(e => e.data && e.name) && action.id;

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

// function removeOperation(
//   operation: TriggerActionOperation
// ) {
//   action[operation] = null;
//   operations.value = operations.value.filter(
//     (e) => e != operation
//   );
// }

function deviceSelected(
  id: string,
  friendlyName: string
) {
  action.id = id;

  // reset
  //action.property = '';
  //action.data = null;
  action.delay = {
    'unit': 'seconds',
    'value': 0,
  };
}

function dataInputChange(value: string) {
  //action.data = value;
}

// function delayInputChange(value: string) {
//   action.delay = parseInt(value);
// }

function exposeSelected(name: string) {
  //action.property = name;

  // reset
  //action.data = null;
  action.delay = {
    'unit': 'seconds',
    'value': 0,
  };
}

function saveAction() {
  if (action.delay) {
    //action.delay = toMillisecs(action.delay);
  }
  emit('save', action);
}
function removeAction() {
  emit('delete', action);
}

function addNewExpose() {
  action.exposes.push({
    name: '',
    data: '',
  });
}
function removeTriggerExpose(expose: AutomationTriggerActionExpose) {
  action.exposes = action.exposes.filter((e) => e != expose);
}

</script>

<template>
  Action {{ action }}
  <ButtonPanel :buttons="buttonPanelItems" />

  <div class="pb-3" />
  <h4>Trigger Action</h4>
  <div class="pb-3" />

  <div class="pb-3">
    <DeviceSelector label="Device to trigger" :id="action.id" @updated="deviceSelected"
      :filter="featureDevicesFilter()" />
  </div>
  <div class="flex align-items-center justify-content-left pb-3">
    <Dropdown :items="dropdownItems" :disabled="action.id == ''" text label="Operations" icon="pi pi-plus"
      size="small" />
  </div>
  <!-- <div v-for="expose in action.exposes" :key="expose.name">
    <div class="col sm:col-4">
      <ExposeSelector :id="action.id" label="Expose" @updated="exposeSelected" :value="expose.name"
        :filter="featureExposeFilter()" />
    </div>
    <div class="col sm:col-4">
      <ExposeDataInput :show-presets="true" :id="action.id" :name="expose.name" label="Set value"
        @updated="dataInputChange" :disabled="expose.name == ''" :value="expose.data"></ExposeDataInput>
    </div>
  </div> -->


  <DataTable :value="action.exposes" selectionMode="single">
    <Column header="Expose">
      <template #body="slotProps">
        <div class="col sm:col-4">
          <ExposeSelector :id="action.id" label="Expose" @updated="exposeSelected" :value="slotProps.data.name"
            :filter="featureExposeFilter()" />
        </div>
        <!-- <div class="col sm:col-4">
          <ExposeDataInput :id="action.id" :name="slotProps.data.name" label="Set value" @updated="dataInputChange"
            :disabled="slotProps.data.name == ''" :value="slotProps.data.data">
          </ExposeDataInput>
        </div> -->

      </template>
    </Column>
    <Column header="Data">
      <template #body="slotProps">
        data {{ slotProps }}
        <div class="col sm:col-4">
          <ExposeDataInput :id="action.id" :name="slotProps.data.name" label="Set value" @updated="dataInputChange"
            :disabled="slotProps.data.name == ''" :value="slotProps.data.data">
          </ExposeDataInput>
        </div>
      </template>
    </Column>
    <Column class="col-sm-1">
      <template #body="slotProps">
        <Button icon="pi pi-trash" variant="text" rounded @click="removeTriggerExpose(slotProps.data)" />
      </template>
    </Column>
  </DataTable>
  <div class="pt-4 flex align-items-center justify-content-center">
    <Button style="width: 99%" icon="pi pi-plus" label="Add Expose" @click="addNewExpose()" text size="small" />
  </div>

  <!-- <div v-for="operation in operations">
    <div style="
        display: flex;
        align-items: center;
        gap: 0.5rem;
      ">
      <InputBox label=" Delay in minutes" :is-numeric="true" :disabled="action.exposes.length == 0"
        @updated="delayInputChange" :value="action[operation]" />
      <Button icon="pi pi-trash" text iconOnly="true" @click="removeOperation(operation)" />
    </div>
  </div> -->
</template>
