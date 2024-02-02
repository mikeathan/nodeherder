<script setup lang="ts">

import { computed, watch, ref, toRef, toRaw, reactive, PropType } from "vue";

import TriggerCondition from "./TriggerCondition.vue"
import TriggerAction from "./TriggerAction.vue";
import Selector from "../input/Selector.vue"
import { store } from "../../store/index";
import { Device } from "@/types/device";
import { AutomationTrigger, AutomationTriggerAction } from "@/types/automation";
import { clearAction, removeCondition, isValid, insertCondition } from "../../contracts/automations"
import { capitalizeText } from "../../modules/formatters/text.formatter";
import { Props } from "../input/Slider.vue";


const props = defineProps({
    id: { type: String },
    trigger: { type: Object as PropType<AutomationTrigger>, default: {} as AutomationTrigger },
});

const action = ref<AutomationTriggerAction>({} as AutomationTriggerAction)
const actionRef = ref<InstanceType<typeof TriggerAction>>()
const trigger = ref<AutomationTrigger>(props.trigger)

watch(
    () => props.trigger,
    () => {
        action.value = JSON.parse(JSON.stringify(props.trigger.action)) as AutomationTriggerAction;
    }, { immediate: true }
)

const emit = defineEmits(['save', 'delete'])

function save() {

    trigger.value.action.data = action.value.data
    trigger.value.action.friendlyname = action.value.friendlyname
    trigger.value.action.id = action.value.id
    trigger.value.action.property = action.value.property
    trigger.value.action.operation = action.value.operation
    trigger.value.action.delay = action.value.delay
    console.log("after save trigger.action: ", trigger.value.action)
    emit('save', trigger.value)
}

function remove() {
    emit('delete', trigger.value)
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
                            <button type="button" class="btn btn-default btn-number"
                                @click="(e) => insertCondition(trigger)">
                                <span class="fa fa-plus"></span>
                            </button>
                        </h5>
                    </th>
                </tr>
                <tbody v-for="( condition, index ) in   trigger.conditions  " :item="condition">
                    <tr>
                        <th scope="w-25">
                            <TriggerCondition :id="props.id" :name="condition.name" :operator="condition.equality"
                                :data="condition.value" @update:name="newValue => condition.name = newValue"
                                @update:value="newValue => condition.value = newValue"
                                @update:operator="newValue => condition.equality = newValue">
                            </TriggerCondition>
                        </th>
                        <td>
                            broblem here it updates the source object, once we fix action then do same here
                            <span class="fa fa-trash-alt fa-sm" @click="removeCondition(trigger, condition)">
                            </span>
                        </td>
                    </tr>
                </tbody>
                <tr>
                    <th scope="col">
                        <h5>Actions
                            <!-- 
                                NOTE : we will use it when we have more than one Actions 
                                <button v-if="trigger.action.id ==''" type="button" class="btn btn-default btn-number ms-3"
                                @click="addAction()">
                                <span class=" fa fa-plus"></span>
                            </button> -->
                        </h5>
                    </th>
                </tr>
                <tbody>

                    <tr>
                        <th scope="w-25">
                            {{ action }}
                            <TriggerAction :action="action" @update="v => action = v" ref="actionRef">
                            </TriggerAction>
                        </th>
                        <td>
                            <span v-if="action.id != ''" class="fa fa-trash-alt fa-sm" @click="(v) => clearTriggerAction()
                                ">
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
</template>