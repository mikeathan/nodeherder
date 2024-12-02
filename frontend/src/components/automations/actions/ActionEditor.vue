<script setup lang="ts">
import { ref, watch, PropType } from "vue";
import { getActionType, ActionType, } from "@/contracts/automations"
import { AutomationTriggerAction } from "@/types/automation";
import { emitClosePanel } from "@/mixins/useAutomationsEventBus";
import { PanelComponents } from "@/mixins/usePanelComponents";


const emit = defineEmits<{
    (e: 'save', action: AutomationTriggerAction): void,
    (e: 'delete', action: AutomationTriggerAction): void,
}>()

const props = defineProps({

    item: {
        type: Object as PropType<AutomationTriggerAction>,
        default: {} as AutomationTriggerAction,
        required: true
    },
    automationId: {
        type: String,
        default: '',
        required: false
    },
    editMode: {
        type: Boolean,
        defaul: false
    }
});

const currentAction = ref(props.item)
const actionType = ref<ActionType>("TriggerAction");

watch(
    () => props.item,
    () => {
        actionType.value = getActionType(props.item);

    }, { immediate: true }
)

function saveAction(action: AutomationTriggerAction): void {
    currentAction.value = action

    console.log('[DEBUG] ActionEditor -  SaveAction ', action.steps);

    emit('save', currentAction.value);
    emitClosePanel('ActionEditor');
}

function removeAction(action: AutomationTriggerAction): void {
    emit('delete', currentAction.value);
    emitClosePanel('ActionEditor');
}

</script>
<template>
    <component :is="PanelComponents[actionType]" v-bind="{ automationId: props.automationId, action: currentAction }"
        @delete="removeAction" @save="saveAction" />
</template>
@