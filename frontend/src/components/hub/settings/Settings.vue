<script setup lang="ts">
import { computed } from "vue";
import { key, store } from "../../../store/index";
import HistorySettings from './HistorySettings.vue';
import { HistorySettingsType, LoggerSettingsType } from "@/types/settings";
import LoggerSettings from "./LoggerSettings.vue";


NEED fixing - it doesnt load the settings
const historySettings = computed(() => {
    if (!store.getters['appconfig/initialized']() as Boolean) {
        store.dispatch('ws/emit', { event: 'loadAppConfig' });
    }
    return store.getters['appconfig/history']() as HistorySettingsType
});

const loggerSettings = computed(() => {
    if (!store.getters['appconfig/initialized']() as Boolean) {
        store.dispatch('ws/emit', { event: 'loadAppConfig' });
    }
    return store.getters['appconfig/logger']() as LoggerSettingsType
});
</script>

<template>
    <div className="content p-0 p-sm-3">
        <h1>Settings</h1>
        <br>
        <br>
        <HistorySettings :settings="historySettings"></HistorySettings>
        <LoggerSettings :settings="loggerSettings"></LoggerSettings>
    </div>
</template>
