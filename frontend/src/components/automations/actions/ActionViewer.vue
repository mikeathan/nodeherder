<script setup lang="ts">
import { ref, watch, PropType, defineAsyncComponent, computed, onMounted, inject, reactive, onUnmounted } from "vue";
import { getActionType, ActionType, AutomationActionTypes } from "@/contracts/automations"
import { AutomationTriggerAction } from "@/types/automation";
import { EventActions, Events, OpenPanelEvent } from "@/types/events.type";
import { LocalEventBus } from "@/composables/eventBus";

const eventBus = inject(LocalEventBus);
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
    console.log("ActionViewer - DELETE")

    emit('delete', action);

    eventBus!.emit('closePanel', 'ActionViewer');
}

function openEditor(): void {
    eventBus!.emit('openPanel', createActionEditorOpenPanelEvent(currentAction.value));
}

function createActionEditorOpenPanelEvent(action: AutomationTriggerAction): OpenPanelEvent {
    return { owner: 'ActionViewer', name: 'ActionEditor', args: { item: action }, events: {} }
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
