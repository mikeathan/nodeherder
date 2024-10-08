<script setup lang="ts">
import { computed } from "vue";
import { key, store } from "../../../store/index";
import Toggle from "../../input/Toggle.vue"


function enableLogging(enabled: boolean) {
    store.dispatch('ws/emit', {
        event: "enableRemoteLogger", message: {
            enabled: enabled,
        }
    });
    store.dispatch('ws/enableRemoteLogger', true);
}

we need a isRemoteLoggerEnabled function

</script>



<template>
    <div className="content p-0 p-sm-3">
        <Toggle :minimal="false" :value="false" :valueOn="true" :valueoff="false"
            @update="(v: boolean) => enableLogging(v)">
        </Toggle>
    </div>
</template>
