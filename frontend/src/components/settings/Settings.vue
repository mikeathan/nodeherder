<script setup lang="ts">
import { computed } from "vue";
import { RouterLink } from "vue-router";
import { store } from "../../store/index";
import { HistorySettings } from "@/types/settings";
import Toggle from '../input/Toggle.vue';
import InputBox from '../input/InputBox.vue';

// for now just load history
// maybe we load everyhing and the individual device settings are in a combo 
const historySettings = computed(() => {
    if (!store.getters['appconfig/initialized']() as Boolean) {
        store.dispatch('ws/emit', { event: 'loadAppConfig' });
    }

    return store.getters['appconfig/history']()
});

function toggleChanged(propName: any, propValue: any) {
    save(propName, propValue);
}

function inputLostFocus(propName: any, propValue: any) {
    save(propName, propValue);
}

function save(propName: any, propValue: any) {
    if (historySettings.value[propName] != propValue) {
        historySettings.value[propName] = propValue;
        store.dispatch('appconfig/saveHistorySettings', historySettings.value as HistorySettings);
    }
}

</script>

<template>
    <div className="content p-0 p-sm-3">
        <h1>Settings</h1>
        <br>
        <br>
        <h2>History</h2>
        <div class="row border-bottom py-1 w-100 align-items-center" v-for="(value, key) in historySettings" :key="key">
            <dl class="col-12 col-md-3">
                <dt>
                    <strong> {{ key }}</strong>
                </dt>
            </dl>
            <div class="col-md-4">
                <div v-if="typeof value === 'boolean'">
                    <Toggle :minimal="false" :value="value" :valueOn="true" :valueoff="false"
                        @update="(v) => toggleChanged(key, v)">
                    </Toggle>
                </div>
                <div v-else>
                    <InputBox :value="value" :disabled="typeof value !== 'number'"
                        :is-numeric="typeof value === 'number'" @lost-focus="(f) => inputLostFocus(key, f)">
                    </InputBox>
                </div>
            </div>
        </div>
    </div>
</template>
