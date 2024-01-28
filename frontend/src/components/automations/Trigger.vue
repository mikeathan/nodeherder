<script setup lang="ts">

import { computed, watch, ref, PropType } from "vue";

import TriggerCondition from "./TriggerCondition.vue"
import TriggerAction from "./TriggerAction.vue";
import Selector from "../input/Selector.vue"
import { store } from "../../store/index";
import { Device } from "@/types/device";
import { AutomationTrigger, AutomationTriggerCondition } from "@/types/automation";
import { EditableAutomationTrigger } from "../../contracts/automations"

const props = defineProps({
    id: { type: String },
    trigger: { type: Object as PropType<AutomationTrigger> },
});

const trigger = ref<EditableAutomationTrigger>(EditableAutomationTrigger.create())
const emit = defineEmits(['save', 'delete'])

watch(
    () => props.trigger,
    () => {

        const obj: AutomationTrigger = JSON.parse(JSON.stringify(props.trigger))
        trigger.value = EditableAutomationTrigger.createFrom(obj)

    }, { immediate: true }
)

function removeAction(event: Event): void {
    console.log("clear action")
    trigger.value.clearAction()
}

function addCondition(): void {
    trigger.value.addCondition()
}

function removeCondition(condition: AutomationTriggerCondition): void {
    trigger.value.removeCondition(condition)
}

function save() {
    emit('save', trigger.value)
}

function remove() {
    emit('delete', trigger.value.idx)
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
                            Trigger {{ trigger.displayName() }}
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
                            <TriggerAction :id="trigger.action.id" :property="trigger.action.property"
                                :data="trigger.action.data" :delay="trigger.action.delay"
                                :operation="trigger.action.operation"
                                @update:id="(id, name) => trigger.setActionDeviceId(id, name)"
                                @update:property="v => trigger.setActionProperty(v)"
                                @update:data="v => trigger.action.data = v" @update:delay="v => trigger.action.delay = v"
                                @update:operation="v => trigger.action.operation = v">
                            </TriggerAction>
                        </th>
                        <td>
                            <span v-if="trigger.action.id != ''" class="fa fa-trash-alt fa-sm" @click="removeAction">
                            </span>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
        <div class="row  pt-3">
            <div class="row pt-3">
                <div class="col">
                    <button type="button" class="btn btn-light" :disabled="trigger.isValid() == false" @click="save">
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