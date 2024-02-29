<script setup lang="ts">
import { computed, ref, watchEffect, watch, PropType, reactive } from "vue";
import TriggerAction from "../TriggerAction.vue"
import StepAction from "../StepAction.vue"
import { OperationType, resolveObjectOperations } from "../../contracts/operations"
import { clearAction, setDeviceId, getActionType, AutomationActionTypes, EditableActionTrigger } from "../../contracts/automations"
import { store } from "../../store/index";
import { Device, Devices, ExposeType } from "@/types/device";
import { KeyyValuePair } from "@/types/types";
import { AutomationTriggerAction } from "@/types/automation";
import { ExposeTypes } from "@/types/device.type";


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

const actionType = ref("");
const action = reactive({ ...props.item })
const device = computed(() => {
    return store.getters["devices/find"](action.id) as Device;
});
watch(
    () => props.item,
    () => {
        actionType.value = getActionType(props.item);
    }, { immediate: true }
)

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
        ActionType:{{ actionType }} - Item:{{ props.item }}
        <component :is="actionType" />
        <!-- <span v-if="action.id != ''" class="fa fa-trash-alt fa-sm" @click="(v) => clear()
            ">
        </span> -->
    </div>
</template>


<!-- const componentMap = {
    ComponentWebComponentsImage: defineAsyncComponent(() =>
      import('../components/Hero1.vue'),
    ),
    ComponentWebComponentsImage: defineAsyncComponent(() =>
      import('../components/Hero2.vue'),
    ),
    ComponentWebComponentsImageText = defineAsyncComponent(() =>
      import('../components/Hero3.vue'),
    )
  }
  </script>
  
  <template>
    <div v-for="item in result.home.data.attributes.webcomponents" :key="item.id">
      <component :is="componentMap[item.__typename]" />
    </div>
  </template> -->