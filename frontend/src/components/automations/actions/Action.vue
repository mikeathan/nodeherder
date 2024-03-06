<script setup lang="ts">
import { ref, watch, PropType, defineAsyncComponent, computed, onMounted } from "vue";
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
    setEditorView(false);
    emit('save', action);
}

function setEditorView(enable: boolean): void {
    isEditorView.value = enable;
}

function removeAction(action: AutomationTriggerAction): void {
    emit('delete', action);
}

onMounted(() => {
    if (!actionView.value) {
        isEditorView.value = true;
    }
});

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
    <div class="row" v-if="isEditorView == false">
        <div class=" col-sm-11" @click="e => setEditorView(true)">
            {{ actionView }}
        </div>
        <div class="col-sm-1">
            <span class="fa fa-trash-alt fa-sm" @click="e => removeAction(props.item)">
            </span>
        </div>
    </div>
    <div v-else>
        <!-- <div class="row">
            <button type="button" class="btn-close" aria-label="Close" @click="e => setEditorView(false)"></button>
        </div> -->
        <component :is="componentMap[actionType]" v-bind="{ action: props.item }" @delete="removeAction"
            @save="saveAction" />
    </div>
</template>
