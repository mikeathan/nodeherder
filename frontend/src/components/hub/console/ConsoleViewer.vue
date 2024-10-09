<script setup lang="ts">
import { computed, onMounted } from "vue";
import { key, store } from "../../../store/index";
import Toggle from "../../input/Toggle.vue"
import { consoleCleanupService } from "@/services/console-cleanup.service";


onMounted(() => {
    consoleCleanupService.startTimer(store);
});

const isEnabled = computed(() => {
    return store.getters['console/isEnabled']();
});


function enableLogging(enabled: boolean) {
    if (enabled == isEnabled.value) {
        return;
    }
    store.dispatch('console/enableRemoteLogging', enabled);
}

</script>

<template>
    <div className="content p-0 p-sm-3">
        <Toggle :minimal="false" :value="isEnabled" :valueOn="true" :valueoff="false"
            @update="(v: boolean) => enableLogging(v)">
        </Toggle>
    </div>
</template>
