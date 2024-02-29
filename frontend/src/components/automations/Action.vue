<script setup lang="ts">
import { computed, ref, watchEffect, watch, PropType, reactive } from "vue";
import DataInput from "../input/DataInput.vue"
import Selector from "../input/Selector.vue"
import { OperationType, resolveObjectOperations } from "../../contracts/operations"
import { clearAction, setDeviceId, getActionType } from "../../contracts/automations"
import { store } from "../../store/index";
import { Device, Devices, ExposeType } from "@/types/device";
import { KeyyValuePair } from "@/types/types";
import { AutomationTriggerAction } from "@/types/automation";
import { toMillisecs, toMinutes } from '@/modules/formatters/time.formatter'
import { ExposeTypes } from "@/types/device.type";

const emit = defineEmits<{
    (e: 'update', action: AutomationTriggerAction): void,
}>()

const props = defineProps({
    type: {
        type: String,
        default: null
    },
    item: {
        type: Object as PropType<AutomationTriggerAction>,
        default: {} as AutomationTriggerAction,
    },
});


watchEffect(() => actionType.value = props.type);
watch(
    () => props.item,
    () => {
        actionType.value = getActionType(props.item);
    }, { immediate: true }
)

const actionType = ref("");
const action = reactive({ ...props.item })
const device = computed(() => {
    return store.getters["devices/find"](action.id) as Device;
});


const clear = (() => {
    clearAction(action)
    emit('update', action)
})

defineExpose({
    clear,
});


</script>

<template>
    <div class="row">
        <!-- <span v-if="action.id != ''" class="fa fa-trash-alt fa-sm" @click="(v) => clear()
            ">
        </span> -->
    </div>
</template>
