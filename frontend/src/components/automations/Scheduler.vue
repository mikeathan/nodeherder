<script setup lang="ts">
import { ref, watch, PropType } from "vue";
import { getActionType, ActionType, } from "@/contracts/automations"
import { AutomationTriggerAction } from "@/types/automation";
import { emitClosePanel } from "@/mixins/useAutomationsEventBus";
import { PanelComponents } from "@/mixins/usePanelComponents";


const props = defineProps({
    id: String,
});

const emit = defineEmits<{
    (e: 'save', action: AutomationTriggerAction): void,
    (e: 'delete', action: AutomationTriggerAction): void,
}>()



</script>
<template>
    <div>

        ID {{ props.id }}
    </div>
</template>
@