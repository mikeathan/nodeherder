<script setup>
import { useStore } from "vuex";
import { computed } from "vue";
import Editor from "./editor.vue"
import { RouterLink } from "vue-router";

const store = useStore();
const automations = computed(() => {

    if (!store.getters["automations/isInitialized"]) {
        store.dispatch('ws/emit', "loadAutomations");
    }
    return store.getters["automations/items"]
});

</script>

<template>
    <div v-for="(automation, name) in automations" :item="automation">
        <RouterLink :to="`/editor/${automation.name}`">{{
            automation.name
        }}</RouterLink>
    </div>
</template>
