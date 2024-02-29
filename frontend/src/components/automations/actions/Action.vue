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
    (e: 'update', action: AutomationTriggerAction): void,
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

// const clear = (() => {
// })

// defineExpose({
//     clear,
// });

</script>

<template>
    <component :is="componentMap[actionType]" v-bind="{ action: props.item }" />
</template>
