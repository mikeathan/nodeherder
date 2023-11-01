<script setup>
import { useStore } from "vuex";
import { computed, ref } from "vue";

const operators = ref([
    { text: '=', value: '=' },
    { text: '<=', value: '<=' },
    { text: '>=', value: '>=' },
    { text: '>', value: '>' },
    { text: '<', value: '<' }
])
const store = useStore();
const selectedDevice = ref(null)
const devices = computed(() => {
    var devices = store.getters["devices/items"];
    return devices;
})

function addCondition() {
    console.log("add condition");
}

function selectDevice() {
    console.log("selectDevice click ", selectedDevice);
    console.log("exposes: ");
    for (var e in selectedDevice.value.exposes) {
        console.log(e);
    }
    // show panel with exposes to choose from

    // TODO:
    // to fix make combo default to first item
    // click should call function and set the selecteditem
    // 
}

</script>
<template>
    <div class="container-fluid p-0 h-100">
        <div class="card">
            <div class="card-body">
                <h3>Create new Automation</h3>
                <div class="col-3">
                    <div class="d-flex">
                        <select id="deviceSelector" style="text-align:center;" class="form-control"
                            v-model="selectedDevice">
                            <option :value="null">Select device</option>
                            <option v-for="device in devices" :value="device" :key="device.id">
                                {{ device.friendly_name }}
                            </option>
                        </select>
                        <!-- <div class="input-group-btn">
                            <button type="button" class="btn btn-secondary text-nowrap"
                                @click="selectDevice(selectedDevice)">Select
                                device</button>
                        </div> -->

                    </div>
                </div>

                <div v-if="selectedDevice != null">
                    <div class="col-3">
                        <select id="exposeSelector" style="text-align:center;" class="form-control">
                            <option v-for="expose in selectedDevice.exposes">
                                {{ expose.name }}
                            </option>
                        </select>
                    </div>
                    // need component
                    <br>
                    <h4>Condition</h4>
                    <div class="row w-50">
                        <div class="col">
                            <input type="text" class="form-control" placeholder="Condition name"
                                onfocus="this.placeholder = ''" onblur="this.placeholder='Condition name'" />
                        </div>
                        <div class="col">
                            <select id="selectOperators" style="text-align:center;" class="form-control">
                                <option v-for="operator in operators" :value="operator.value" :key="operator.value">
                                    {{ operator.text }}
                                </option>
                            </select>
                        </div>
                        <div class="col">
                            <input type="text" style="text-align:center;" class="form-control" placeholder="Condition value"
                                onfocus="this.placeholder = ''" onblur="this.placeholder='Condition value'" />
                        </div>
                        <div class="col">
                            <input type="button" class="btn btn-secondary text-nowrap" value="Add"
                                @click="addCondition()" />
                        </div>

                    </div>
                </div>

            </div>
        </div>
    </div>
</template>
