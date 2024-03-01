<script setup lang="ts">
import { ref, watch, PropType, defineAsyncComponent } from "vue";
import { getActionType, ActionType } from "@/contracts/automations"
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

</script>

<template>
    need to have simple view for existing actions once we click it i can focus the view in a panel for editing

    can we display object here in simple view without loading component ?
    once clicked hide trigger panel and show only action. maybe it needs to be done in trigger ?
    <component :is="componentMap[actionType]" v-bind="{ action: props.item }" @delete="removeAction" @save="saveAction" />
</template>
