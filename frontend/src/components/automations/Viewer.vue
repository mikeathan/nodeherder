<script setup>
import { useStore } from "vuex";
import { computed } from "vue";

import { RouterLink } from "vue-router";

const store = useStore();
const automations = computed(() => {

    if (!store.getters["automations/isInitialized"]) {
        store.dispatch('ws/emit', { event: "loadAutomations" });
    }
    return store.getters["automations/items"]
});

function onDeleteAutomationClick(id) {
    // emit delete event
    store.dispatch('ws/emit', {
        event: "deleteAutomation", message: {
            id: id,
        }
    });
    console.log("delete automation: Id", id);
}

function saveAutomation(id) {
    // emit save event
    var values = Object.values(automations.value);
    var res = values.filter(k => k.id == id);
    if (res.length != 0) {
        store.dispatch('automations/save', res[0]);
        console.log("save automation: Id", id);
    }
}
</script>

<template>
    <div className="content p-0 p-sm-3">

        <table class="table responsive table-hover">
            <thead>
                <tr>
                    <th scope="col">#</th>
                    <th scope="col">Name</th>
                    <th scope="col">Description</th>
                    <th scope="col">Enabled</th>

                    <th scope="col"></th>
                </tr>
            </thead>
            <tbody>
                <tr v-for="(automation, name, index) in automations" :item="automation">
                    <th scope="row">{{ index + 1 }}</th>
                    <td>
                        <RouterLink :to="`/editor/${automation.id}`">{{
                            automation.friendlyname
                        }}</RouterLink>
                    </td>
                    <td>
                        {{ automation.description }}
                    </td>
                    <td>
                        <div class=" form-check form-switch">
                            <label class="form-check-label">Enable</label>

                            <input class="form-check-input" type="checkbox" role="switch" id="flexSwitchCheckDefault"
                                v-model="automation.enabled" @change="saveAutomation(automation.id)">
                        </div>
                    </td>
                    <td>
                        <span class="fa fa-trash-alt fa-lg" @click="onDeleteAutomationClick(automation.id)"></span>
                    </td>
                </tr>
            </tbody>
        </table>

        <div>

            <RouterLink :to="`/creator`">
                Create automations
            </RouterLink>
        </div>
    </div>
</template>
