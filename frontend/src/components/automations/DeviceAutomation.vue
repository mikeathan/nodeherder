<script setup>
import { useStore } from "vuex";
import { computed, watch, ref, onBeforeMount, watchEffect } from "vue";

import { ExposeTrigger, DeviceTrigger, Operators } from "../../models/automation"
import Conditions from "./Conditions"
import TriggerCondition from "./TriggerCondition"

const props = defineProps({
    id: String,
    trigger: Object
});

const store = useStore();
const exposes = ref({})
const deviceTrigger = ref(new DeviceTrigger(props.id))
const deviceExposes = ref({})
const conditions = ref([]);
const selectedExpose = ref("")
const device = computed(() => {
    return store.getters["devices/find"](props.id);
});


function exposeSelectionChanged(event) {

    var value = event.target.value;
    if (value == "") {
        //selectedExpose.value = null; // maybe refactor ????
        return;
    }


    // if (exposes.value[value] != null) {
    //     console.log(value, " exists with conditions: ", exposes.value[value].Conditions.length)
    //     return
    // }

    // exposes.value[value] = new ExposeTrigger(value)
    // deviceExposes[props.id].triggers.push(exposes.value[value])
    // console.log("expose added: ", value, " : ", exposes.value[value], "conditions: ", exposes.value[value].Conditions.length)
}
watch(
    () => props.id,
    (t) => {
        console.log("id changed ", props.id)
        selectedExpose.value = null
        if (deviceExposes[props.id] == null) {
            deviceExposes[props.id] = new DeviceTrigger(props.id)
            console.log("new device trigger")
        }
    },
    { immediate: true }
);


function add(event) {
    console.log("add ", event)
    conditions.value.push(event);
}

function remove(event) {
    console.log("remove ", event)

    var index = conditions.value.findIndex(item => item.Id === event);
    if (index != -1) {
        conditions.value.splice(index, 1);
    }
}

</script>
<template>
    <div class="container-fluid p-0 h-100" v-if="device != null"> <!-- to fix condition-->

        <div class="col-3">
            <select id="exposeSelector" style="text-align:center;" class="form-control" @change="exposeSelectionChanged"
                v-model="selectedExpose">
                <option :value="null">Select expose</option>
                <option v-for="expose in device.exposes" :value="expose.name" :key="expose.name">
                    {{ expose.name }}
                </option>
            </select>
        </div>
        <br>
        Device: {{ props.id }} - {{ selectedExpose }}

        <div>
            <div class="row w-50" v-if="selectedExpose != null">
                <TriggerCondition :id="0" :name="selectedExpose" :operator="Operators[0].value" :data="''"
                    @add="add($event)">
                </TriggerCondition>
            </div>

            <div v-for="condition in conditions">
                <div class="row w-50">
                    <TriggerCondition :id="condition.Id" :name="condition.Name" :operator="condition.Operator"
                        :key="condition.id" :data="condition.Data" @remove="remove($event)"></TriggerCondition>
                </div>
            </div>
            <div class="col-50">
                <div class="btn-group">
                    <button type="button" class="btn btn-default" :disabled="conditions.length == 0">
                        Save <!-- add conditions to trigger, and reenable options drop down-->
                    </button>
                    <button type="button" class="btn btn-default">
                        Cancel <!-- send event to enable back options dropdown-->
                    </button>
                </div>
            </div>

        </div>
    </div>
</template>
