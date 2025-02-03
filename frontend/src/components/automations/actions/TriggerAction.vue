<script setup lang="ts">
  import { computed, ref, watch, PropType, reactive } from 'vue';
  import { AutomationTriggerAction, AutomationTriggerActionExpose } from '@/types/automation';
  import { toMillisecs, toMinutes } from '@/modules/formatters/time.formatter';
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

  const action = ref<AutomationTriggerAction>({} as AutomationTriggerAction);

  const operations = ref<TriggerActionOperation[]>([]);

  watch(
    () => props.action,
    () => {
      console.log('trigger action changed', props.action);
      action.value = JSON.parse(JSON.stringify(props.action)) as AutomationTriggerAction;
      if (props.action.delay) {
        // action.delay = toMinutes(props.action.delay);
        // operations.value.push('delay');
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
    createTriggerActionOperatorsDropdowitems((e: TriggerActionOperation) => addOperation(e))
  );

  function addOperation(operation: TriggerActionOperation) {
    operations.value.push(operation);
  }

  function addNewExpose() {
    const expose: AutomationTriggerActionExpose = {
      name: '',
      data: null,
    };
    action.value.exposes.push(expose);
  }

  function removeOperation(
    operation: TriggerActionOperation
  ) {
    //action[operation] = null;
    operations.value = operations.value.filter(
      (e) => e != operation
    );
  }

  function deviceSelected(id: string, friendlyName: string) {
    action.value.id = id;

    // reset
    //action.property = '';
    //action.data = null;
    action.value.delay = {
      unit: 'seconds',
      value: 0,
    };
  }

  function dataInputChange(expose: AutomationTriggerActionExpose, value: string) {
    expose.data = value;
  }

  function delayInputChange(value: string) {
    //action.delay = parseInt(value);
  }

  function exposeSelected(expose: AutomationTriggerActionExpose, name: string) {
    expose.name = name;

    // reset
    //action.data = null;
    // action.value.delay = {
    //   unit: 'seconds',
    //   value: 0,
    // };
  }

  function saveAction() {
    if (action.value.delay) {
      //action.delay = toMillisecs(action.delay);
    }
    emit('save', action.value);
  }
  function removeAction() {
    emit('delete', action.value);
  }

  function removeTriggerExpose(expose: AutomationTriggerActionExpose) {
    action.value.exposes = action.value.exposes.filter((e) => e != expose);
  }
</script>

<template>
  Action {{ action }}
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
  <div class="flex align-items-center justify-content-left pb-3">
    <Dropdown
      :items="dropdownItems"
      :disabled="action.id == ''"
      text
      label="Operations"
      icon="pi pi-plus"
      size="small" />
  </div>
  <div v-for="expose in action.exposes" :key="expose.name">
    <div class="row">
      <div class="col sm:col-4">
        <ExposeSelector
          :id="action.id"
          label="Expose"
          @updated="(e) => exposeSelected(expose, e)"
          :value="expose.name"
          :filter="featureExposeFilter()" />
      </div>
      <div class="col sm:col-4">
        <ExposeDataInput
          :show-presets="true"
          :id="action.id"
          :name="expose.name"
          label="Set value"
          @updated="(e) => dataInputChange(expose, e)"
          :disabled="expose.name == ''"
          :value="expose.data"></ExposeDataInput>
      </div>
      <div class="col sm:col-2">
        <Button icon="pi pi-trash" variant="text" rounded @click="removeTriggerExpose(expose)" />
      </div>
    </div>
  </div>

 
  <div class="pt-4 flex align-items-center justify-content-center">
    <Button style="width: 99%" icon="pi pi-plus" label="Add Expose" @click="addNewExpose()" text size="small" />
  </div>

  <div v-for="operation in operations">
    <div style="
        display: flex;
        align-items: center;
        gap: 0.5rem;
      ">
      <InputBox label=" Delay in minutes" :is-numeric="true" :disabled="action.exposes.length == 0"
        @updated="delayInputChange" :value="action[operation]" />
        
        <!-- <InputBox :label="interval.unit" :value="interval.value" :is-numeric="true"
        @lost-focus="(f) => inputLostFocus(key, f)" /> -->
      <Button icon="pi pi-trash" text iconOnly="true" @click="removeOperation(operation)" />
    </div>
  </div>
</template>
