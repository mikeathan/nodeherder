<script setup>
import { useStore } from "vuex";
import { ref, watch, computed } from "vue";
import Trigger from "./Trigger"
import { ExposeTrigger, DeviceTrigger } from "../../models/automation"

import DataInput from "../input/DataInput.vue"

import { useRouter } from 'vue-router'

const props = defineProps({
    id: String,
});

const store = useStore();
const router = useRouter()
const automation = ref(new DeviceTrigger())
const selectedExpose = ref("")
const selectedTrigger = ref(null)


watch(
    () => props.id,
    () => {
        var sourceAutomation = store.getters["automations/find"](props.id);
        if (sourceAutomation != undefined) {
            // make a deep copy to make it not reactive
            automation.value = JSON.parse(JSON.stringify(sourceAutomation))
        }
    }, { immediate: true }
)
const device = computed(() => {
    return store.getters["devices/find"](props.id);
});

function isSaveEnabled() {
    var values = automation.value.triggers.filter(k => k.action != null);
    return values.length == automation.value.triggers.length
}

function addNewTrigger() {
    var trigger = new ExposeTrigger('')
    automation.value.triggers.push(trigger)
    selectedTrigger.value = trigger
}

function save() {
    store.dispatch('automations/save', automation.value);
    router.push("/viewer")
}

function saveTrigger(event) {
    // that wont work for updates
    // work only for adding new triggers

    console.log("save trigger ", event)
    //automation.triggers.push(event)
    selectedTrigger.value = null;
}


function getConditionsDescription(trigger) {
    var conditions = trigger.conditions
    if (conditions.length == 0) {
        return ""
    }
    var condition = conditions[0];
    var description = condition.name + " " + condition.equality + " " + condition.value;
    if (conditions.length > 1) {
        description += "..."
    }

    return description;
}
function getActionDescription(trigger) {
    if (trigger.action == null) {
        return "<EMPTY>"
    }

    var description = trigger.action.friendlyname + " " +
        trigger.action.property;

    return description;
}

function rowClicked(trigger) {
    selectedTrigger.value = trigger;
}

function onDeleteTriggerClick(event, triggerId) {
    automation.value.triggers.splice(triggerId, 1);
}


</script>
<style scoped>
.custom-control-input {
    transform: scale(1.4);
}
</style>
<template>
    <div v-if="automation">
        <div class="container-fluid p-0 h-100">

            <div class="card col-xl-5 col-md-6 col-sm-3">
                <div class="card-header ">
                    <div class="pt-3 ">
                        <label class="form-check-label">Id</label>
                        <DataInput :data="props.id" alignment="left" :disabled="true">
                        </DataInput>
                    </div>
                    <div class="pt-3">
                        <label class="form-check-label">Friendly Name</label>
                        <DataInput :data="automation.friendlyname" alignment="left" type="string" :disabled="true">
                        </DataInput>
                    </div>
                    <div class="pt-3  pb-4">
                        <label class="form-check-label">Description</label>
                        <DataInput :data="automation.description" @update:data="(value) => automation.description = value"
                            type="string" alignment="left">
                        </DataInput>
                    </div>

                    <div class="pb-3">
                        <div class=" form-check form-switch ms-2">
                            <label class="form-check-label ms-3">Enabled</label>

                            <input class="form-check-input custom-control-input" type="checkbox" role="switch"
                                id="flexSwitchCheckDefault" v-model="automation.enabled">
                        </div>
                    </div>

                    <div class="btn-group">
                        <button type="button" class="btn btn-light" @click="save" :disabled="isSaveEnabled() == false">
                            Save
                        </button>
                        <router-link :to="`/viewer`" tag="span">
                            <button type="button" class="btn btn-light">
                                Cancel
                            </button> </router-link>
                    </div>
                </div>

                <div class="card-body ">
                    <table class="table responsive table-hover " v-if="selectedTrigger == null">
                        <thead>
                            <tr>
                                <th scope="col">#</th>
                                <th scope="col">Action</th>
                                <th scope="col">Conditions</th>
                                <th scope="col">
                                    <button type="button" class="btn btn-default btn-number" @click="addNewTrigger()">
                                        <span class="fa fa-plus"></span>
                                    </button>
                                </th>
                            </tr>
                        </thead>
                        <tbody v-for="(trigger, index) in automation.triggers" :item="trigger">
                            <tr>
                                <th scope="row">
                                    {{ index + 1 }}
                                </th>
                                <td @click="rowClicked(trigger)">
                                    {{ getActionDescription(trigger) }}
                                </td>
                                <td>
                                    {{ getConditionsDescription(trigger) }}
                                </td>
                                <td>
                                    <span class="fa fa-trash-alt fa-sm" @click="onDeleteTriggerClick($event, index)"
                                        data-bs-toggle="collapse" data-bs-target>
                                    </span>
                                </td>
                            </tr>
                        </tbody>
                    </table>

                    <div class="row" v-else>
                        <button type="button" class="btn-close" aria-label="Close"
                            @click="() => selectedTrigger = null"></button>
                        <div class="col">
                            <Trigger :id="props.id" :trigger="selectedTrigger" @save="saveTrigger">
                            </Trigger>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>
