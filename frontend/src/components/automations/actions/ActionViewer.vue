<script setup lang="ts">
import { ref, watch, PropType, computed, h } from "vue";
import { getActionType, ActionType, AutomationActionTypes } from "@/contracts/automations"
import { AutomationTriggerAction } from "@/types/automation";
import { EventActions, OpenPanelEvent } from "@/types/events.type";
import { emitOpenPanel } from "@/mixins/useAutomationsEventBus";

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
    },
    automationId: {
        type: String,
        default: '',
        required: false
    },
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

            // eg. Attic light brightness = (brightness + action_time * 0.3) 
            return currentAction.value.friendlyname + " =  (" + stepView + ")";
        } else if (actionType.value == AutomationActionTypes.PresetRotation) {

            return "[" + currentAction.value.friendlyname + "] preset rotate [" + currentAction.value.property + "]";
        } else if (actionType.value == AutomationActionTypes.Trigger) {

            // "Delay 5 min" // new row
            // const test = "Set Attic Light state off WITH 5 min Delay";

            // const h3 = h("h3", "Set")
            // const device = h("div", { color: "gray" }, "Attic Light")
            // const property = h("div", "state off")
            // const delayText = h("div", "Delay 5 min");
            // use h() to build elements
            //return test;
            const delay = currentAction.value.delay ? `[Delay = ${currentAction.value.delay}]` : "";
            return `[${currentAction.value.friendlyname}] > [${currentAction.value.property} =  ${currentAction.value.data}] ${delay}`
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


// function removeAction(action: AutomationTriggerAction): void {
//     emit('delete', action);
//     emitClosePanel('ActionViewer');
// }

function openEditor(): void {
    emitOpenPanel(createActionEditorOpenPanelEvent(currentAction.value));
}

function createActionEditorOpenPanelEvent(action: AutomationTriggerAction): OpenPanelEvent {
    return { name: 'ActionEditor', args: { automationId: props.automationId, item: action }, events: props.editEvents }
}

</script>

<template>
    <div class="row">
        <div class=" col-sm-11" @click="openEditor()">
            {{ actionView }}
        </div>
    </div>
</template>