<script setup>
import { useStore } from "vuex";
import { computed } from "vue";

import { RouterLink, useRouter } from "vue-router";

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
function navigate() {
    console.log("navigate")
    var creatorPath = useRouter().push("/creator");
    console.log(creatorPath)
}
const creatorRoute = computed(() => {
    return useRouter().push("/creator");

});
</script>

<template>
    <div className="content p-0 p-sm-3">

        <table class="table responsive table-hover">
            <thead>
                <tr>
                    <th scope="col">#</th>
                    <th scope="col">Name</th>
                    <th scope="col">Description</th>
                    <th scope="col"></th>
                </tr>
            </thead>
            <tbody>
                <tr v-for="(automation, name, index) in automations" :item="automation">
                    <th scope="row">{{ index + 1 }}</th>
                    <td>
                        <RouterLink :to="`/editor/${automation.id}`">{{
                            automation.friendlyName
                        }}</RouterLink>
                    </td>
                    <td>
                        {{ automation.description }}
                    </td>

                    <td>
                        <span class="fa fa-trash-alt fa-lg" @click="onDeleteAutomationClick(automation.id)"></span>
                    </td>
                </tr>
            </tbody>
        </table>

        <div>
            <button type="button" class="btn btn-primary mt-3" @click="navigate()">Create automations</button>
        </div>
    </div>
</template>
