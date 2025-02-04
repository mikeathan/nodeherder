<script setup lang="ts">
import { ref, watch, PropType, computed, h } from 'vue';
import { ActionType, AutomationActionTypes, isPresetRotationAction, isStepAction, isTriggerAction } from '@/contracts/automations';
import { AutomationAction, AutomationTriggerAction, AutomationTriggerActionExpose } from '@/types/automation';
import { EventActions, OpenPanelEvent } from '@/types/events.type';
import { emitOpenPanel } from '@/mixins/useAutomationsEventBus';
import { toMinutes } from '@/modules/formatters/time.formatter';
import { store } from '@/store';
import { Device } from '@/types/device';
import { transformPresetCyclingAction, transformStepAction, transformTriggerAction } from '@/transformers/automation/action-transformers';

const emit = defineEmits<{
  (e: 'delete', action: AutomationTriggerAction): void;
  (e: 'edit', events: EventActions): void;
}>();

const props = defineProps({
  item: {
    type: Object as PropType<AutomationAction>,
    default: {} as AutomationAction,
    required: true,
  },
  editEvents: {
    type: Object as PropType<EventActions>,
    default: {} as EventActions,
    required: true,
  },
  automationId: {
    type: String,
    default: '',
    required: false,
  },
});

const currentAction = ref(props.item);
const actionType = ref<ActionType>('TriggerAction');

const deviceNameFromId = (currentAction: AutomationAction): string => {
  const device = store.getters['hub/findDevice'](currentAction.id) as Device;
  if (device == undefined) {
    return '';
  }
  return device.friendly_name;
};

const actionView = computed(() => {
  if (currentAction.value.id) {
    const friendlyName = deviceNameFromId(currentAction.value);

    if (isTriggerAction(currentAction.value)) {

      return transformTriggerAction(friendlyName, currentAction.value);
    }
    if (isPresetRotationAction(currentAction.value)) {
      return transformPresetCyclingAction(friendlyName, currentAction.value)
    }

    if (isStepAction(currentAction.value)) {
      return transformStepAction(friendlyName, currentAction.value);
    }

    //   switch (actionType.value) {
    //     case AutomationActionTypes.Trigger:
    //       return [
    //         `Set the ${deviceNameFromId(currentAction.value)} ${currentAction.value.property} to ${currentAction.value.data} `,
    //         currentAction.value.delay
    //           ? `in ${toMinutes(
    //             currentAction.value.delay
    //           )} minutes`
    //           : null,
    //       ];
    //     case AutomationActionTypes.PresetRotation:
    //       return [
    //         `Rotate the ${deviceNameFromId(currentAction.value)} ${currentAction.value.property}`,
    //       ];
    //     case AutomationActionTypes.Step:
    //       let stepValue = '';
    //       currentAction.value.steps.forEach((step) => {
    //         stepValue +=
    //           step.property + ' ' + step.operator + ' ';
    //       });
    //       stepValue += currentAction.value.data;
    //       return [
    //         `Adjusting ${deviceNameFromId(currentAction.value)} ${currentAction.value.property}`,
    //         `by [${stepValue}] steps`,
    //       ];

    //     // TODO; appy styles eg = <p>Adjusting <span class="highlight-word">attic light</span> <span class="highlight-word">brightness</span> by [action-time + 10] steps.</p>
  }
  return [];
});

watch(
  () => props.item,
  () => {
    actionType.value = props.item.type;
  },
  { immediate: true }
);

function openEditor(): void {
  emitOpenPanel(createActionEditorOpenPanelEvent(currentAction.value));
}

function createActionEditorOpenPanelEvent(action: AutomationAction): OpenPanelEvent {
  return {
    name: 'ActionEditor',
    args: {
      automationId: props.automationId,
      item: action,
    },
    events: props.editEvents,
  };
}
</script>
<template>
  <div class="" @click="openEditor()">
    <Card>
      <template #content>
        <div v-for="item in actionView">
          <p v-html="item" />
        </div>
      </template>
    </Card>
  </div>
</template>
