<script setup lang="ts">
  import { computed, PropType, reactive } from 'vue';
  import { AutomationTriggerAction } from '@/types/automation';
  import ButtonPanel from '@/components/controls/ButtonPanel.vue';
  import { createSaveDeleteButtonItems } from '../../../configs/automation/trigger-dropdown.config';
  import DeviceSelector from '@/components/controls/DeviceSelector.vue';
  import ExposeSelector from '@/components/controls/ExposeSelector.vue';
  import {
    presetsDevicesFilter,
    presetExposeFilter,
  } from '@/configs/automation/device.config';

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

  const action = reactive({ ...props.action });
  const buttonPanelItems = computed(() => {
    const actionIsValid = action.property && action.id;
    return createSaveDeleteButtonItems(
      () => saveAction(),
      () => removeAction(),
      !actionIsValid,
      !actionIsValid
    );
  });

  const emit = defineEmits<{
    (e: 'save', action: AutomationTriggerAction): void;
    (e: 'delete', action: AutomationTriggerAction): void;
  }>();

  function deviceSelected(
    deviceId: string,
    friendlyName: string
  ) {
    action.id = deviceId;
    action.friendlyname = friendlyName;
    action.property = '';
    action.steps = [];
  }

  function exposeSelected(value: string) {
    action.property = value;
  }

  function saveAction(): void {
    emit('save', action);
  }

  function removeAction(): void {
    emit('delete', action);
  }
</script>
<template>
  <div class="row pb-3">
    <div class="col">
      <ButtonPanel
        :buttons="buttonPanelItems"></ButtonPanel>
    </div>
  </div>

  <h4>Rotate Action</h4>
  <div class="pb-3" />

  <div class="row pb-3">
    <DeviceSelector
      label="Device to trigger"
      :id="action.id"
      @updated="deviceSelected"
      :filter="presetsDevicesFilter()">
    </DeviceSelector>
  </div>
  <div class="row pb-3">
    <ExposeSelector
      label="Expose"
      :id="action.id"
      :value="action.property"
      @updated="exposeSelected"
      :filter="presetExposeFilter()">
    </ExposeSelector>
  </div>
</template>
