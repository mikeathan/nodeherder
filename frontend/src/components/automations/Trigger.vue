<script setup lang="ts">

import { computed, watch, ref, toRef, toRaw, reactive, PropType } from "vue";

import TriggerCondition from "./TriggerCondition.vue"
import TriggerAction from "./TriggerAction.vue";
import Selector from "../input/Selector.vue"
import { store } from "../../store/index";
import { Device } from "@/types/device";
import { AutomationTrigger, AutomationTriggerAction, AutomationTriggerCondition } from "@/types/automation";
import { EditableAutomationTrigger, EditableTriggerCondition, clearAction } from "../../contracts/automations"
import { capitalizeText } from "../../modules/formatters/text.formatter";


const props = defineProps({
    id: { type: String },
    trigger: { type: Object as PropType<AutomationTrigger>, default: {} as AutomationTrigger },
});

// TODO: maybe use some automation context responsible for these operations 
const action = ref<AutomationTriggerAction>({} as AutomationTriggerAction)
const trigger = ref<AutomationTrigger>(props.trigger)

const emit = defineEmits(['save', 'delete'])

function addCondition(): void {
    trigger.value.conditions.push(new EditableTriggerCondition());
}

function removeCondition(condition: AutomationTriggerCondition): void {
    // trigger.value.removeCondition(condition)
}

function save() {
    trigger.value.action.data = action.value.data
    trigger.value.action.friendlyname = action.value.friendlyname
    trigger.value.action.id = action.value.id
    trigger.value.action.property = action.value.property
    trigger.value.action.operation = action.value.operation
    trigger.value.action.delay = action.value.delay

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
                            <button type="button" class="btn btn-default btn-number" @click="addCondition()">
                                <span class="fa fa-plus"></span>
                            </button>
                        </h5>
                    </th>
                </tr>
                <tbody v-for="(condition, index) in  trigger.conditions " :item="condition">
                    <tr>
                        <th scope="w-25">
                            <TriggerCondition :id="props.id" :name="condition.name" :operator="condition.equality"
                                :data="condition.value" @update:name="newValue => condition.name = newValue"
                                @update:value="newValue => condition.value = newValue"
                                @update:operator="newValue => condition.equality = newValue">
                            </TriggerCondition>
                        </th>
                        <td>
                            <span class="fa fa-trash-alt fa-sm" @click="removeCondition(condition)">
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

                            <TriggerAction :action="trigger.action" @update="v => action = v">
                            </TriggerAction>
                        </th>
                        <td>
                            <span v-if="trigger.action.id != ''" class="fa fa-trash-alt fa-sm" @click="(v) => clearAction(trigger.action)
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
                        <!-- :disabled="trigger.isValid() == false"  -->
                        Save
                    </button>
                    <button type="button" class="btn btn-light" @click="remove">
                        Delete
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>