<script setup>
import { useStore } from "vuex";
import { computed, watch, ref, onBeforeMount } from "vue";

import { ExposeTrigger, DeviceTrigger, Operators } from "../../models/automation"
import Conditions from "./Conditions"
import TriggerCondition from "./TriggerCondition"

const props = defineProps({
    id: String
});

const store = useStore();
const exposes = ref({})
const deviceTrigger = ref(new DeviceTrigger(props.id))
const deviceExposes = ref({})
const conditions = ref([]);
const selectedExpose = ref(null)
const device = computed(() => {
    return store.getters["devices/find"](props.id);
});

function exposeSelectionChanged(event) {

    var value = event.target.value;
    if (value == "") {
        //selectedExpose.value = null; // maybe refactor ????
        return;
    }

    var deviceTrigger = deviceExposes[props.id]
    var index = deviceTrigger.triggers.findIndex(item => item.Name === value);
    var expose;
    if (index != 0) {
        expose = exposes.Conditions[index]
    }
    else {
        expose = new ExposeTrigger(value)
        device.triggers.push(expose)
    }

    conditions.value = expose.Conditions;
    console.log("expose selected ", value)
    selectedExpose.value = event.target.value; // maybe refactor ????
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
        // conditions.value = ex
    },
    { immediate: true }
);
// Device Trigger
//  Triggers
// {
//    Expose
//    {
//         Conditions
//         {
//
//         }
//         Actions
//         {
//    
//         }
//     }
// }

</script>
<template>
    <div v-if="device != null" class="container-fluid p-0 h-100">

        <div class="col-3">
            <select id="exposeSelector" style="text-align:center;" class="form-control" @change="exposeSelectionChanged">
                <option value="">Select expose</option>
                <option v-for="expose in device.exposes" :value="expose.name" :key="expose.name">
                    {{ expose.name }}
                </option>
            </select>
        </div>
        <br>
        Device: {{ props.id }} - {{ selectedExpose }}

        <div>
            <!-- <Expose :id="props.id" :expose="exposes[selectedExpose]"></Expose> -->
            <div class="row w-50">
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

            <!-- <h4>Condition</h4>
            <div v-if="selectedExpose != null" class="col">
                <Conditions :expose="exposes[selectedExpose]">
                </Conditions>
            </div> -->
        </div>
    </div>
    <div v-else>
    </div>
</template>
