<script setup>
import { useStore } from "vuex";
import { computed, watch, ref, watchEffect } from "vue";

import { Condition, ActionTrigger } from "../../models/automation"
import TriggerCondition from "./TriggerCondition"
import TriggerAction from "./TriggerAction.vue";
import Selector from "../input/Selector.vue"

const props = defineProps({
    id: String,
    trigger: Object,
});

const store = useStore();
const trigger = ref(null)
const emit = defineEmits(['save', 'delete'])

watch(
    () => props.trigger,
    () => {

        trigger.value = JSON.parse(JSON.stringify(props.trigger))
        trigger.value.conditions.forEach(function callback(condition, index) {
            condition.idx = index + 1
        });

    }, { immediate: true }
)

const device = computed(() => {
    return store.getters["devices/find"](props.id);
});

function addAction() {
    var newAction = new ActionTrigger()
    trigger.value.action = newAction;
}

function removeAction(event) {
    trigger.value.action = null;
}

function addCondition() {
    var condition = new Condition()
    trigger.value.conditions.push(condition)
}

function removeCondition(event) {
    var index = trigger.value.conditions.filter(k => k.idx != event);
    if (index != -1) {
        trigger.value.conditions.splice(index, 1);
    }
}

function isSaveEnabled() {
    return trigger.value.name != "" && trigger.value.action != null;
}

function actionIdUpdated(id, friendlyName) {
    trigger.value.action.id = id
    trigger.value.action.friendlyname = friendlyName
}

function save() {
    emit('save', trigger.value)
}

function remove() {
    emit('delete', trigger.value.idx)
}

function exposesList() {
    // todo:
    //var result = Object.keys(obj).map((key) => [key, obj[key]]);
    var list = {}
    for (const [key, expose] of Object.entries(device.value.exposes)) {
        list[expose.name] = expose.name
    }
    return list
}

function capitalize(val) {
    return val.charAt(0).toUpperCase() + val.slice(1);
}
</script>

<template>
    <div class="container-fluid p-0 h-100">
        <!-- TODO:  -->
        <!-- if automation for device exists message user else we overwrite it -->

        <div class="row" v-if="trigger.name == ''">

            <Selector placeholder="Select trigger" :items="exposesList()" :value="trigger.name" alignment="left"
                :disabled="props.trigger.name != ''" @update:data="v => trigger.name = v">
            </Selector>
        </div>
        <div v-else>
            <h5> Trigger {{ capitalize(trigger.name) }} </h5>

            <div class="row pt-3">

                <!-- Conditions -->

                <table class="table">
                    <thead>
                        <tr>
                            <th scope="col">
                                <h5>Conditions
                                    <button type="button" class="btn btn-default btn-number" @click="addCondition()">
                                        <span class="fa fa-plus"></span>
                                    </button>
                                </h5>
                            </th>
                            <th scope="col">#</th>
                        </tr>
                    </thead>
                    <tbody v-for="(condition, index) in  trigger.conditions " :item="condition">
                        <tr>
                            <th scope="w-25">
                                <TriggerCondition :id="props.id" :index="condition.idx" :name="condition.name"
                                    :operator="condition.equality" :key="condition.idx" :data="condition.value"
                                    @update:name="newValue => condition.name = newValue"
                                    @update:value="newValue => condition.value = newValue"
                                    @update:operator="newValue => condition.equality = newValue">
                                </TriggerCondition>
                            </th>
                            <td>
                                <span class="fa fa-trash-alt fa-sm" @click="removeCondition(condition.index)">
                                </span>
                            </td>
                        </tr>
                    </tbody>

                    <!-- <tbody>
                        <tr>
                            <th scope="w-25">
                                <TriggerAction v-if="trigger.action != null" :id="trigger.action.id"
                                    :property="trigger.action.property" :data="trigger.action.data"
                                    :delay="trigger.action.delay" :step="trigger.action.step" @update:id="actionIdUpdated"
                                    @update:property="newValue => trigger.action.property = newValue"
                                    @update:data="newValue => trigger.action.data = newValue"
                                    @update:delay="newValue => trigger.action.delay = newValue"
                                    @update:step="newValue => trigger.action.step = newValue">
                                </TriggerAction>
                            </th>
                            <td>
                                <span v-if="trigger.action != null" class="fa fa-trash-alt fa-sm" @click="removeAction">
                                </span>
                            </td>
                        </tr>
                    </tbody> -->
                </table>
            </div>


            <!-- Actions -->
            <div class="row  pt-3">
                <table class="table">
                    <thead>
                        <tr>
                            <th scope="col">
                                <h5>Actions
                                    <button v-if="trigger.action == null" type="button" class="btn btn-default btn-number"
                                        @click="addAction()">
                                        <span class=" fa fa-plus"></span>
                                    </button>
                                </h5>
                            </th>
                            <th scope="col">#</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr>
                            <th scope="w-25">
                                <TriggerAction v-if="trigger.action != null" :id="trigger.action.id"
                                    :property="trigger.action.property" :data="trigger.action.data"
                                    :delay="trigger.action.delay" :step="trigger.action.step" @update:id="actionIdUpdated"
                                    @update:property="newValue => trigger.action.property = newValue"
                                    @update:data="newValue => trigger.action.data = newValue"
                                    @update:delay="newValue => trigger.action.delay = newValue"
                                    @update:step="newValue => trigger.action.step = newValue">
                                </TriggerAction>
                            </th>
                            <td>
                                <span v-if="trigger.action != null" class="fa fa-trash-alt fa-sm" @click="removeAction">
                                </span>
                            </td>
                        </tr>
                    </tbody>
                </table>


                <div class="row pt-3">
                    <div class="col">
                        <button type="button" class="btn btn-light" :disabled="isSaveEnabled() == false" @click="save">
                            Save
                        </button>
                        <button type="button" class="btn btn-light" @click="remove">
                            Delete
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>