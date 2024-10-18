<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { key, store } from '../../../store/index';
import Toggle from '../../input/Toggle.vue';
import { consoleCleanupService } from '@/services/console-cleanup.service';
import { LogMessageType } from '@/types/console.type';
import { formatTimestamp } from '@/utils/date.utils';
import { LoggerSettingsType } from '@/types/settings';
import { getConsoleLevelClass } from '@/contracts/console';

onMounted(() => {
    consoleCleanupService.startTimer(store);
});

const loggerSettings = computed(() => {
    if (!store.getters['appconfig/initialized']() as Boolean) {
        store.dispatch('ws/emit', { event: 'loadAppConfig' });
    }
    const set = store.getters['appconfig/logger']()
    if (set == undefined) {
        return {} as LoggerSettingsType
    }
    return store.getters['appconfig/logger']() as LoggerSettingsType
});

const messages = computed(() => {
    return store.getters[
        'console/messages'
    ]() as LogMessageType[];
});


function enableLogging(enabled: boolean) {
    if (enabled == loggerSettings.value.enableRemoteLogger) {
        return;
    }
    loggerSettings.value.enableRemoteLogger = enabled;
    store.dispatch('appconfig/saveLoggerSettings', loggerSettings.value);
}

function clearConsole() {
    store.commit('console/clear');
}
</script>

<style></style>
<template>
    <div className="content p-0 p-sm-3">
        <div class="card">
            <div class="card-header">
                <label class=" pe-1">Remote logger:</label>
                <Toggle :minimal="true" :value="loggerSettings.enableRemoteLogger" :valueOn="true" :valueoff="false"
                    @update="(v: boolean) => enableLogging(v)">
                </Toggle>

                <button class="btn btn-link" @click="clearConsole">Clear</button>
            </div>

            <div class="card-body">
                <div v-for="(message, index) in messages" :key="message.timestamp">
                    <span style="width: 60px;" :class="`badge ${getConsoleLevelClass(message.level)}`">{{ message.level
                        }}</span>
                    &nbsp;
                    <small class="pe-1">{{
                        formatTimestamp(message.timestamp)
                        }}</small>
                    &nbsp;
                    <code>{{ message.message }}</code>
                </div>
            </div>
        </div>
    </div>
</template>
