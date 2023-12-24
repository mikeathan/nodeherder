<script setup>
import { useStore } from "vuex";
import { ref, watch } from "vue";
import Trigger from "./Trigger"
import { ExposeTrigger, DeviceTrigger } from "../../models/automation"
import DataInput from "../input/DataInput.vue"
import { useRouter } from 'vue-router'

const emit = defineEmits(['cancel'])

const props = defineProps({
    id: String,
});

const store = useStore();
const router = useRouter()
const automation = ref(new DeviceTrigger())
const selectedTrigger = ref(null)


watch(
    () => props.id,
    () => {
        var sourceAutomation = store.getters["automations/find"](props.id);
        if (sourceAutomation != undefined) {
            // make a deep copy to make it not reactive
            automation.value = JSON.parse(JSON.stringify(sourceAutomation))
            automation.value.triggers.forEach(function callback(trigger, index) {
                trigger.idx = index
            });
        } else {

            automation.value = new DeviceTrigger()
            var device = store.getters["devices/find"](props.id);
            if (device != undefined) {
                automation.value.id = device.id
                automation.value.friendlyname = device.friendly_name
            }
        }
    }, { immediate: true }
)


function isSaveEnabled() {
    if (automation.value.triggers.length == 0) {
        return false
    }

    var values = automation.value.triggers.filter(k => k.action != null);
    return values.length == automation.value.triggers.length
}

function createNewTrigger() {
    selectedTrigger.value = new ExposeTrigger('')
}

function cancel() {
    emit('cancel')
}

function saveAutomation(trigger) {
    store.dispatch('automations/save', automation.value);
    router.push("/viewer")
}

function deleteAutomation() {
    var sourceAutomation = store.getters["automations/find"](props.id);
    if (sourceAutomation != undefined) {
        console.log(automation.value);
        store.dispatch('automations/delete', automation.value.id);
        // todo; alert message box to ask user
        router.push("/viewer")
    }
}

function deleteTrigger(triggerIdx) {
    if (triggerIdx |= -1) {
        automation.value.triggers.splice(triggerIdx, 1);
    }
    selectedTrigger.value = null// close trigger panel
}

function saveTrigger(trigger) {
    if (trigger.idx == -1) {
        automation.value.triggers.push(trigger)
        automation.value.triggers.forEach(function callback(trigger, index) {
            trigger.idx = index
        });
    } else {
        automation.value.triggers[trigger.idx] = trigger
    }

    selectedTrigger.value = null; // close trigger panel
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
    deleteTrigger(triggerId)
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
                        <button type="button" class="btn btn-light" @click="saveAutomation"
                            :disabled="isSaveEnabled() == false">
                            Save
                        </button>
                        <button type="button" class="btn btn-light" @click="deleteAutomation">
                            Delete
                        </button>
                        <button type="button" class="btn btn-light" @click="cancel">
                            Cancel
                        </button>
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
                                    <button type="button" class="btn btn-default btn-number" @click="createNewTrigger()">
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
                            <Trigger :id="props.id" :trigger="selectedTrigger" @save="saveTrigger" @delete="deleteTrigger">
                            </Trigger>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>
