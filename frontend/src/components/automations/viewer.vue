<script setup>
import { useStore } from "vuex";
import { computed, onMounted } from "vue";

import { wssend } from "@/services/ws";

const store = useStore();
const automations = computed(() => store.getters.automations);
onMounted(() => {
    wssend("loadAutomations", "")
});
</script>

<template>
    {{ automations }}
</template>
