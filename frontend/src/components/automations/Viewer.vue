<script setup lang="ts">
import { computed } from "vue";
import { RouterLink } from "vue-router";
import { store } from "../../store/index";
import { Automations } from "@/types/automation";
import AutomationStatus from "./schedule/AutomationStatus.vue";

const automations = computed(() => {

    if (!store.getters["automations/initialized"]() as Boolean) {
        store.dispatch('ws/emit', { event: "loadAutomations" });
    }
    return store.getters["automations/listAll"]() as Automations
});

function onDeleteAutomationClick(id: string): void {
    // emit delete event
    store.dispatch('ws/emit', {
        event: "deleteAutomation", message: {
            id: id,
        }
    });
}

function saveAutomation(id: string): void {
    // emit save event
    var values = Object.values(automations.value).filter(k => k.id == id);
    if (values.length != 0) {
        store.dispatch('automations/save', values[0]);
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
                <tr v-for="(automation, index) in automations" :item="automation">
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
                        <AutomationStatus :automation="automation" />
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
