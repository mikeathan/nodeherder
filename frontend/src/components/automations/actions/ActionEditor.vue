<script setup lang="ts">
import { ref, watch, PropType, defineAsyncComponent, inject } from "vue";
import { getActionType, ActionType, } from "@/contracts/automations"
import { AutomationTriggerAction } from "@/types/automation";
import { LocalEventBus } from "@/composables/eventBus";


const eventBus = inject(LocalEventBus);

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

watch(
    () => props.item,
    () => {
        actionType.value = getActionType(props.item);

    }, { immediate: true }
)

function saveAction(action: AutomationTriggerAction): void {
    currentAction.value = action
    emit('save', currentAction.value);

    eventBus!.emit('closePanel', 'ActionEditor');
}

function removeAction(action: AutomationTriggerAction): void {
    emit('delete', currentAction.value);

    eventBus!.emit('closePanel', 'ActionEditor');
}

</script>

<template>
    <div>
        <component :is="componentMap[actionType]" v-bind="{ action: currentAction }" @delete="removeAction"
            @save="saveAction" />
    </div>
</template>
