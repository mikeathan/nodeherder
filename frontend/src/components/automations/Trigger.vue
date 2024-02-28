<script setup lang="ts">

import { computed, watch, ref, PropType, toRef } from "vue";
import TriggerCondition from "./TriggerCondition.vue"
import TriggerAction from "./TriggerAction.vue";
import Selector from "../input/Selector.vue"
import { store } from "../../store/index";
import { Device } from "@/types/device";
import { AutomationTrigger, AutomationTriggerAction, AutomationTriggerCondition, AutomationTriggerConditions } from "@/types/automation";
import { isValid, EditableTriggerCondition } from "../../contracts/automations"
import { capitalizeText } from "../../modules/formatters/text.formatter";
import ActionSelectorDialog from "../dialogs/ActionSelectorDialog.vue"

const props = defineProps({
    id: { type: String },
    trigger: { type: Object as PropType<AutomationTrigger>, default: {} as AutomationTrigger },
});

// TODO: can be refactor to some automation context
const conditions = ref<AutomationTriggerConditions>({} as AutomationTriggerConditions)
const action = ref<AutomationTriggerAction>({} as AutomationTriggerAction)
const actionRef = ref<InstanceType<typeof TriggerAction>>()
const trigger = ref<AutomationTrigger>(props.trigger)
const showDialog = ref<boolean>(false);

watch(
    () => props.trigger,
    () => {
        action.value = JSON.parse(JSON.stringify(props.trigger.action)) as AutomationTriggerAction;
        conditions.value = JSON.parse(JSON.stringify(props.trigger.conditions)) as AutomationTriggerConditions;
    }, { immediate: true }
)

const emit = defineEmits(['save', 'delete'])
function save() {

    trigger.value.conditions = conditions.value
    trigger.value.action = action.value

    emit('save', trigger.value)
}

function remove() {
    emit('delete', trigger.value)
}

function addCondition() {
    conditions.value.push(new EditableTriggerCondition());
}

function removeTriggerCondition(condition: AutomationTriggerCondition) {
    conditions.value = conditions.value.filter((c) => c != condition);
}

const exposesList = computed(() => {
    const device = store.getters["devices/find"](props.id) as Device;
    return Object.assign({}, ...Object
        .values(device.exposes)
        .map((e) => ({ [e.name]: e.name })))
})

function clearTriggerAction() {
    actionRef.value?.clear()
}

</script>

<template>
    <div class="container-fluid p-0 h-100">
        <!-- TODO:  -->
        <!-- if automation for device exists message user else we overwrite it -->

        <div class="row" v-if="trigger.name == ''">
            <Selector placeholder="Select trigger" :items="exposesList" :value="trigger.name" alignment="left"
                :disabled="trigger.name != ''" @update:data="v => trigger.name = v">
            </Selector>
        </div>
        <div class="row" v-else>

            <!-- Conditions -->
            <table class="table">
                <thead>
                    <tr>
                        <th scope="col">
                            Trigger {{ capitalizeText(trigger.name) }}
                        </th>
                        <th scope="col">#</th>
                    </tr>
                </thead>
                <tr>
                    <th scope="col">
                        <h5>Conditions
                            <button type="button" class="btn btn-default btn-number" @click="(e) => addCondition()">
                                <span class="fa fa-plus"></span>
                            </button>
                        </h5>
                    </th>
                </tr>
                <tbody v-for="( condition, index ) in   conditions  " :item="condition">
                    <tr>
                        <th scope="w-25">
                            <TriggerCondition :id="props.id" :name="condition.name" :operator="condition.equality"
                                :data="condition.value" @update:name="newValue => condition.name = newValue"
                                @update:value="newValue => condition.value = newValue"
                                @update:operator="newValue => condition.equality = newValue">
                            </TriggerCondition>
                        </th>
                        <td>
                            <span class="fa fa-trash-alt fa-sm" @click="removeTriggerCondition(condition)">
                            </span>
                        </td>
                    </tr>
                </tbody>
                <tr>
                    <th scope="col">
                        <h5>Actions

                            <button v-if="trigger.action.id == ''" type="button" class="btn btn-default btn-number ms-3"
                                @click="showDialog = true">
                                <span class=" fa fa-plus"></span>
                            </button>
                        </h5>
                    </th>
                </tr>
                <tbody>

                    <tr>
                        <th scope="w-25">
                            <TriggerAction :action="action" @update="v => action = v" ref="actionRef">
                            </TriggerAction>
                        </th>
                        <td>
                            <span v-if="action.id != ''" class="fa fa-trash-alt fa-sm" @click="(v) => clearTriggerAction()">
                            </span>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
        <div class="row  pt-3">
            <div class="row pt-3">
                <div class="col">
                    <button type="button" class="btn btn-light" @click="save">
                        Save
                    </button>
                    <button type="button" class="btn btn-light" :disabled="isValid(trigger) == false" @click="remove">
                        Delete
                    </button>
                </div>
            </div>
        </div>
    </div>
    <ActionSelectorDialog :show="showDialog" @close="e => showDialog = e">
    </ActionSelectorDialog>
</template>