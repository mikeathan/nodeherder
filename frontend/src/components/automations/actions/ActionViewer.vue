<script setup lang="ts">
import { ref, watch, PropType, computed, h } from "vue";
import { getActionType, ActionType, AutomationActionTypes } from "@/contracts/automations"
import { AutomationTriggerAction } from "@/types/automation";
import { EventActions, OpenPanelEvent } from "@/types/events.type";
import { emitOpenPanel } from "@/mixins/useAutomationsEventBus";
import { toMinutes } from "@/modules/formatters/time.formatter";

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

    // step action type
    if (currentAction.value.id) {
        switch (actionType.value) {
            case AutomationActionTypes.Trigger:
                return [
                    `SET ${currentAction.value.friendlyname} ${currentAction.value.property} ${currentAction.value.data}`,
                    currentAction.value.delay ?
                        `Delay ${toMinutes(currentAction.value.delay)} min` : null
                ]
            case AutomationActionTypes.PresetRotation:
                return [
                    `SET ${currentAction.value.friendlyname}`,
                    `ROTATE ${currentAction.value.property}`
                ]
            case AutomationActionTypes.Step:
                let stepValue = '';
                currentAction.value.steps.forEach(step => {
                    stepValue += step.property + " " + step.operator + " ";
                });
                stepValue += currentAction.value.data
                return [
                    `SET ${currentAction.value.friendlyname} ${currentAction.value.property}`,
                    `STEP ${stepValue}`
                ]
        }
    }
    return [];
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

<style>
.list-group-item {
    color: gray;
}

/* .list-group-item.list-group-item:hover {
    background-color: gray;
} */
</style>
<template>
    <div class="row">
        <div class=" col-sm-11" @click="openEditor()">
            <div class="card" style="width: 20rem;">
                <ul class="list-group list-group-flush">
                    <li class="list-group-item" v-for="item in actionView">
                        {{ item }}
                    </li>
                </ul>
            </div>
        </div>
    </div>
</template>