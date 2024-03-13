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
    editMode: {
        type: Boolean,
        defaul: false
    }
});

const currentAction = ref(props.item)
const actionType = ref<ActionType>("TriggerAction");
const editMode = ref<boolean>(props.editMode);

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

const isEditMode = computed(() => {
    return actionView.value == '' || editMode.value;
});

watch(
    () => props.item,
    () => {
        actionType.value = getActionType(props.item);

    }, { immediate: true }
)

function saveAction(action: AutomationTriggerAction): void {

    currentAction.value = action
    //setEditorView(false);
    emit('save', currentAction.value);

    emitter.emit('closePanel', 'Action');

    // can we send event to close panel here 
}

// closepanel event received from: Action Panel.vue:15:12
// close component:  Action  history previous size:  2  size 1 Panel.vue:71:12
// closepanel event received from: Action Panel.vue:15:12
// close component:  Trigger  history previous size:  1  size 0 Panel.vue:71:12
// no more components. emit panel close to parent Panel.vue:75:16


function removeAction(action: AutomationTriggerAction): void {
    emit('delete', action);

    emitter.emit('closePanel', 'Action');
}

ONCE WE DELETE EXISTING STEP ACTION WE TRIGGER THE DELETE CLOSE PANEL EVENT TWICE
function createActionOpenPanelEvent(action: AutomationTriggerAction, editMode: boolean): OpenPanelEvent {
    const events: EventActions = {
        'delete': () => { removeAction(action) },
        'save': () => { saveAction(action) },
    };

    return { name: 'Action', args: { item: action, editMode: editMode }, events: events }
}

function enableEditorView(): void {
    console.log("Action  emit openpanel")
    emitter.emit('openPanel', createActionOpenPanelEvent(currentAction.value, true));
}


onMounted(() => {
    // if (!actionView) {
    //     isEditorView.value = true;
    // }
});


</script>

<template>
    <div class="row" v-if="isEditMode == false">
        <div class=" col-sm-11" @click="enableEditorView()">
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
