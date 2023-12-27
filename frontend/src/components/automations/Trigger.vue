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
const selectedAction = ref(null)
const trigger = ref(null)

const emit = defineEmits(['save', 'delete'])

watch(
    () => props.trigger,
    () => {
        selectedAction.value = ""
        // select action options to current action if not null
        if (props.trigger.action != null) {
            selectedAction.value = props.trigger.action.id
        }
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

function add() {

    // newAction.delay = delay.value
    // newAction.step = step.value
    // newAction.data = data.value
    // newAction.friendlyname = device.value.friendly_name
    // newAction.id = device.value.id
    // newAction.property = property.value
    // newAction.type = device.value.exposes[property.value].type
    //emit("add", newAction)
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

</script>

<template>
    <div class="container-fluid p-0 h-100">
        <!-- TODO:  -->
        <!-- mobile dimensions are wrong -->
        <!-- fix triggeractonnew - check refactoring logic -->
        <!-- fix triggeractonnew - props dont update unless we remove and add again -->
        <!-- check url design above for styling of creator text input -->
        <!-- if automation for device exists message user else we overwrite it -->

        <div class="row">
            <h5>Trigger</h5>
            <Selector placeholder="Select trigger" :items="exposesList()" :value="trigger.name" alignment="left"
                :disabled="props.trigger.name != ''" @update:data="v => trigger.name = v">
            </Selector>
        </div>
        <br>
        <!-- Conditions -->
        <div class="col-md-5">
            <h5>Conditions
                <button type="button" class="btn btn-default btn-number" @click="addCondition()">
                    <span class="fa fa-plus"></span>
                </button>
            </h5>
        </div>
        <div v-for="condition in  trigger.conditions ">
            <div class="row">
                <TriggerCondition :id="props.id" :index="condition.idx" :name="condition.name"
                    :operator="condition.equality" :key="condition.idx" :data="condition.value"
                    @remove="removeCondition($event)" @update:name="newValue => condition.name = newValue"
                    @update:value="newValue => condition.value = newValue"
                    @update:operator="newValue => condition.equality = newValue">
                </TriggerCondition>
            </div>
        </div>

        <!-- Actions -->
        <div class="col-md-5">
            <h5>
                Action
                <button type="button" class="btn btn-default btn-number" @click="addAction">
                    <span class="fa fa-plus"></span>
                </button>
            </h5>

        </div>
        <!-- existing action -->
        <div v-if="trigger.action != null" class="row">
            <TriggerAction :id="trigger.action.id" :property="trigger.action.property" :data="trigger.action.data"
                :delay="trigger.action.delay" :step="trigger.action.step" @remove="removeAction"
                @update:data="newValue => trigger.action.data = newValue"
                @update:delay="newValue => trigger.action.delay = newValue"
                @update:step="newValue => trigger.action.step = newValue">
            </TriggerAction>
        </div>
        <div class="row">
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
</template>