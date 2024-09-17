<script setup lang="ts">
import { computed } from "vue";
import { key, store } from "../../store/index";
import InputBox from '../input/InputBox.vue';
import HistorySettings from './HistorySettings.vue';
import { HistorySettingsType } from "@/types/settings";

const historySettings = computed(() => {
    if (!store.getters['appconfig/initialized']() as Boolean) {
        store.dispatch('ws/emit', { event: 'loadAppConfig' });
    }
    return store.getters['appconfig/history']() as HistorySettingsType
});

</script>

<template>
    <div className="content p-0 p-sm-3">
        <h1>Settings</h1>
        <br>
        <br>
        <HistorySettings :settings="historySettings"></HistorySettings>
    </div>
</template>
