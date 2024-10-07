<script setup lang="ts">
import { computed } from "vue";
import { key, store } from "../../../store/index";
import HistorySettings from './HistorySettings.vue';
import { HistorySettingsType } from "@/types/settings";

function enableLogging(enabled: boolean) {
    store.dispatch('ws/emit', {
        event: "enableRemoteLogger", message: {
            enabled: enabled,
        }
    });
    store.dispatch('ws/enableRemoteLogger', true);
}

</script>


<template>
    <div className="content p-0 p-sm-3">
        <Toggle :minimal="false" :value="value" :valueOn="true" :valueoff="false"
            @update="(v: boolean) => enableLogging(v)">
        </Toggle>
    </div>
</template>
