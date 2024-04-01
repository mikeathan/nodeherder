<script setup lang="ts">
import { ref, watch, PropType, computed } from "vue";
import { getActionType, ActionType, AutomationActionTypes } from "@/contracts/automations"
import { AutomationTriggerAction } from "@/types/automation";
import { EventActions, OpenPanelEvent } from "@/types/events.type";
import { emitClosePanel, emitOpenPanel } from "@/mixins/useAutomationsEventBus";

const emit = defineEmits<{
    (e: 'delete', action: AutomationTriggerAction): void,
    (e: 'edit', events: EventActions): void,
}>()

const props = defineProps({
    item: {
        type: Object as PropType<AutomationTriggerAction>,
        default: {} as AutomationTriggerAction,
        required: true
    },
    editEvents: {
        type: Object as PropType<EventActions>,
        default: {} as EventActions,
        required: true
    }
});


const currentAction = ref(props.item)
const actionType = ref<ActionType>("TriggerAction");

const actionView = computed(() => {

    let stepView = "";
    // step action type
    if (currentAction.value.id) {
        if (actionType.value == AutomationActionTypes.Step) {
            currentAction.value.steps.forEach(step => {
                stepView += step.property + " " + step.operator + " ";
            });

            stepView += currentAction.value.data

            // eg. Attic lightbrightness = (brightness + action_time * 0.3) 
            return currentAction.value.friendlyname + " =  (" + stepView + ")";
        } else if (actionType.value == AutomationActionTypes.PresetRotation) {
            return "[" + currentAction.value.friendlyname + "] preset rotate [" + currentAction.value.property + "]";
        }

    }

    return stepView;
});


watch(
    () => props.item,
    () => {
        actionType.value = getActionType(props.item);

    }, { immediate: true }
)


function removeAction(action: AutomationTriggerAction): void {
    emit('delete', action);
    emitClosePanel('ActionViewer');
}

function openEditor(): void {
    emitOpenPanel(createActionEditorOpenPanelEvent(currentAction.value));
}

function createActionEditorOpenPanelEvent(action: AutomationTriggerAction): OpenPanelEvent {
    return { name: 'ActionEditor', args: { item: action }, events: props.editEvents }
}

</script>

<template>
    <div class="row">
        <div class=" col-sm-11" @click="openEditor()">
            {{ actionView }}
        </div>
        <div class="col-sm-1">
            <span class="fa fa-trash-alt fa-sm" @click="e => removeAction(currentAction)">
            </span>
        </div>
    </div>
</template>