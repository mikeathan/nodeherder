<script setup>
import { useStore } from "vuex";
import { computed } from "vue";
import Editor from "./editor.vue"
import { RouterLink } from "vue-router";

const store = useStore();
const automations = computed(() => {

    if (!store.getters["automations/isInitialized"]) {
        store.dispatch('ws/emit', { event: "loadAutomations" });
    }
    return store.getters["automations/items"]
});

</script>

<template>
    <table class="table responsive table-hover">
        <thead>
            <tr>
                <th scope="col">#</th>
                <th scope="col">Name</th>
                <th scope="col">Description</th>
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
            </tr>
        </tbody>
    </table>
</template>
