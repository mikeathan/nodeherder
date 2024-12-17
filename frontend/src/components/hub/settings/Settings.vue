<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { store } from "../../../store/index";
import HistorySettings from './HistorySettings.vue';
import { HistorySettingsType, LoggerSettingsType } from "@/types/settings";
import LoggerSettings from "./LoggerSettings.vue";

const historyConfig = ref<HistorySettingsType>();
const loggerConfig = ref<LoggerSettingsType>();

const historySettings = computed(() => {
    if (!store.getters['appconfig/initialized']() as Boolean) {
        store.dispatch('ws/emit', { event: 'loadAppConfig' });
    }
    return historyConfig.value = store.getters['appconfig/history']() as HistorySettingsType
});

const loggerSettings = computed(() => {
    if (!store.getters['appconfig/initialized']() as Boolean) {
        store.dispatch('ws/emit', { event: 'loadAppConfig' });
    }
    return loggerConfig.value = store.getters['appconfig/logger']() as LoggerSettingsType
});

</script>

<template>
    <Card>
        <template #title>
            <h2>Settings</h2>
        </template>
        <template #content>
            <HistorySettings :settings="historySettings"></HistorySettings>
            <LoggerSettings :settings="loggerSettings"></LoggerSettings>
        </template>
    </Card>
</template>
