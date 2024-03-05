<script setup lang="ts">
import { ref, watch, PropType, defineAsyncComponent, computed } from "vue";
import { getActionType, ActionType, AutomationActionTypes } from "@/contracts/automations"
import { AutomationTriggerAction } from "@/types/automation";

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

const actionType = ref<ActionType>("TriggerAction");
const isEditorView = ref<boolean>(false);


watch(
    () => props.item,
    () => {
        actionType.value = getActionType(props.item);
    }, { immediate: true }
)

function saveAction(action: AutomationTriggerAction): void {
    emit('save', action);
}

function removeAction(action: AutomationTriggerAction): void {
    emit('delete', action);
}

const actionView = computed(() => {

    let stepView = "";
    // step action type
    const action = props.item;
    if (action.id) {
        if (actionType.value == AutomationActionTypes.Step) {
            action.steps.forEach(step => {
                stepView += step.property + " " + step.operator + " ";
            });

            stepView += action.data

            // eg. Attic lightbrightness = (brightness + action_time * 0.3) 
            return action.friendlyname + " =  (" + stepView + ")";
        }
    }

    return stepView;
});
</script>

<template>
    <div class="row" v-if="actionView != ''">
        <div class=" col-sm-11">
            {{ actionView }} double click to open editor
        </div>
        <div class="col-sm-1">
            <span class="fa fa-trash-alt fa-sm" @click="e => removeAction(props.item)">
            </span>
        </div>
    </div>
    <div v-else> if its isEditorView then show
        <component :is="componentMap[actionType]" v-bind="{ action: props.item }" @delete="removeAction"
            @save="saveAction" />
    </div>
</template>
