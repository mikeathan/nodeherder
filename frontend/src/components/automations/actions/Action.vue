<script setup lang="ts">
import { ref, watch, PropType, defineAsyncComponent, computed, onMounted, inject, reactive } from "vue";
import { getActionType, ActionType, AutomationActionTypes } from "@/contracts/automations"
import { AutomationTriggerAction } from "@/types/automation";
import { Emitter } from 'mitt'
import { EventActions, Events, OpenPanelEvent } from "@/types/events.type";

const emitter = inject('emitter') as Emitter<Events>;

const componentMap = {
    "TriggerAction": defineAsyncComponent(() =>
        import("./TriggerAction.vue"),
    ),
    "StepAction": defineAsyncComponent(() =>
        import("./StepAction.vue"),
    ),
}

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

function saveAction(action: AutomationTriggerAction): void {

    currentAction.value = action
    setEditorView(false);
    emit('save', currentAction.value);

    todo!!!!!!
    // can we send event to close panel here 
}

function createActionOpenPanelEvent(action: AutomationTriggerAction): OpenPanelEvent {
    const events: EventActions = {
        'delete': () => { removeAction(action) },
        'save': () => { saveAction(action) },
    };

    return { name: 'Trigger', args: { action: action }, events: events }
}

function setEditorView(enable: boolean): void {

    //isEditorView.value = enable;
    //console.log("Action  emit openpanel")
    //emitter.emit('openPanel', createActionOpenPanelEvent(currentAction.value)); //// TESTING
}

function removeAction(action: AutomationTriggerAction): void {
    emit('delete', action);
}

onMounted(() => {
    // if (!actionView) {
    //     isEditorView.value = true;
    // }
});


</script>

<template>
    <br>
    <div class="row" v-if="actionView != ''">
        <div class=" col-sm-11" @click="e => setEditorView(true)">
            {{ actionView }}
        </div>
        <div class="col-sm-1">
            <span class="fa fa-trash-alt fa-sm" @click="e => removeAction(currentAction)">
            </span>
        </div>
    </div>
    <div v-else>
        <!-- <div class="row">
            <button type="button" class="btn-close" aria-label="Close" @click="e => setEditorView(false)"></button>
        </div> -->
        <component :is="componentMap[actionType]" v-bind="{ action: currentAction }" @delete="removeAction"
            @save="saveAction" />
    </div>
</template>
