<script setup>
import { useStore } from "vuex";
import { computed, onBeforeMount } from "vue";
import Automation from "./automation.vue"


const store = useStore();
const automations = computed(() => {

    if (!store.getters["automations/isInitialized"]) {
        store.dispatch('ws/emit', "loadAutomations");
    }
    return store.getters["automations/items"]
});

onBeforeMount(() => {
});

</script>

<template>
    <p>

        <Automation v-for="automation in automations" :item="automation" :key="automation.name"></Automation>
    </p>
</template>
