<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { key, store } from '../../../store/index';
import Toggle from '../../input/Toggle.vue';
import { consoleCleanupService } from '@/services/console-cleanup.service';
import { LogMessageType } from '@/types/event-logs.type';
import { formatTimestamp } from '@/utils/date.utils';
import { LoggerSettingsType } from '@/types/settings';

const consoleDiv = ref<HTMLDivElement>();

onMounted(() => {
    consoleCleanupService.startTimer(store);
});

const loggerSettings = computed(() => {
    if (!store.getters['appconfig/initialized']() as Boolean) {
        store.dispatch('ws/emit', { event: 'loadAppConfig' });
    }
    return store.getters['appconfig/logger']() as LoggerSettingsType
});

const messages = computed(() => {
    return store.getters[
        'console/messages'
    ]() as LogMessageType[];
});

watch(
    () => messages,
    () => {
        // if (consoleDiv.value) {
        //     consoleDiv.value.scrollIntoView({ behavior: 'smooth' });
        // }

    }, { deep: true }
)

function enableLogging(enabled: boolean) {
    if (enabled == loggerSettings.value.enableRemoteLogger) {
        return;
    }
    loggerSettings.value.enableRemoteLogger = enabled;
    store.dispatch('appconfig/saveLoggerSettings', loggerSettings.value);
}

</script>

<style>
.index {
    margin-right: 20px;
}

.timestamp {
    margin-right: 20px;
}

.level {
    margin-right: 20px;
}
</style>
<template>
    <div className="content p-0 p-sm-3">

        <Toggle :minimal="false" :value="loggerSettings.enableRemoteLogger" :valueOn="true" :valueoff="false"
            @update="(v: boolean) => enableLogging(v)">
        </Toggle>


    </div>

    <div ref="consoleDiv">
        <div v-for="(message, index) in messages" :key="message.timestamp">
            <span class="index">{{ index + 1 }}</span>
            <span class="timestamp">{{
                formatTimestamp(message.timestamp)
                }}</span>
            <span class="level">{{ message.level }}</span>
            <span class="message">{{ message.message }}</span>
        </div>
    </div>
</template>
