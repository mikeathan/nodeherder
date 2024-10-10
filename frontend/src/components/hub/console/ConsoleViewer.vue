<script setup lang="ts">
import { computed, onMounted } from "vue";
import { key, store } from "../../../store/index";
import Toggle from "../../input/Toggle.vue"
import { consoleCleanupService } from "@/services/console-cleanup.service";
import { LogMessageType } from "@/types/event-logs.type";
import { formatTimestamp } from "@/utils/date.utils";


onMounted(() => {
    consoleCleanupService.startTimer(store);
});

const isEnabled = computed(() => {
    return store.getters['console/isEnabled']();
});

const messages = computed(() => {
    return store.getters['console/messages']() as LogMessageType[];
});


function enableLogging(enabled: boolean) {
    if (enabled == isEnabled.value) {
        return;
    }
    store.dispatch('console/enableRemoteLogging', enabled);
}

</script>

<style>
ul {
    list-style: none;
    padding: 0;
}

li {
    display: flex;
    align-items: center;
    padding: 0px;
}

.index {
    margin-right: 10px;
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
        <Toggle :minimal="false" :value="isEnabled" :valueOn="true" :valueoff="false"
            @update="(v: boolean) => enableLogging(v)">
        </Toggle>

        <div className="console-messages">
            <ul>
                <li v-for="(message, index) in messages" :key="message.timestamp">
                    <span class="index">{{ index + 1 }}</span>
                    <span class="timestamp">{{ formatTimestamp(message.timestamp) }}</span>
                    <span class="level">{{ message.level }}</span>
                    <span class="message">{{ message.message }}</span>
                </li>
            </ul>
        </div>
    </div>
</template>
