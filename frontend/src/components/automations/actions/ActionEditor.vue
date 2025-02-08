<script setup lang="ts">
import { ref, watch, PropType } from 'vue';
import { ActionType, AutomationAction, AutomationActionTypes } from '@/types/automation.type.js';
import { emitClosePanel } from '@/mixins/useAutomationsEventBus';
import { PanelComponents } from '@/mixins/usePanelComponents';

const emit = defineEmits<{
  (e: 'save', action: AutomationAction): void;
  (e: 'delete', action: AutomationAction): void;
}>();

const props = defineProps({
  item: {
    type: Object as PropType<AutomationAction>,
    default: {} as AutomationAction,
    required: true,
  },
  automationId: {
    type: String,
    default: '',
    required: false,
  },
  editMode: {
    type: Boolean,
    defaul: false,
  },
});

const currentAction = ref(props.item);
const actionType = ref<ActionType>(AutomationActionTypes.Trigger);

watch(
  () => props.item,
  () => {
    actionType.value = props.item.type;
  },
  { immediate: true }
);

function saveAction(action: AutomationAction): void {
  currentAction.value = action;

  emit('save', currentAction.value);
  emitClosePanel('ActionEditor');
}

function removeAction(action: AutomationAction): void {
  emit('delete', currentAction.value);
  emitClosePanel('ActionEditor');
}
</script>
<template>
  <component :is="PanelComponents[actionType]" v-bind="{
    automationId: props.automationId,
    action: currentAction,
  }" @delete="removeAction" @save="saveAction" />
</template>
@
