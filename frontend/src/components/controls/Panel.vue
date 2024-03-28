<script setup lang="ts">
import { ref, watch, PropType, defineAsyncComponent, computed, onMounted, inject, onUnmounted, provide, InjectionKey } from "vue";
import { Emitter } from 'mitt'
import { EventActions, Events, OpenPanelEvent } from "@/types/events.type";

import { useMittEvent, useMittEvents, LocalEventBus, EventHandlers } from "@/composables/eventBus";
import mitt from "mitt";
import { KeyyValuePair } from "@/types/types";


type PanelKey = string
type Map = { [key: PanelKey]: any }

const componentMap: Map = {
    "Trigger": defineAsyncComponent(() =>
        import("../automations/Trigger.vue"),
    ),
    "TriggerAction": defineAsyncComponent(() =>
        import("../automations/actions/TriggerAction.vue"),
    ),
    "StepAction": defineAsyncComponent(() =>
        import("../automations/actions/StepAction.vue"),
    ),
    "ActionEditor": defineAsyncComponent(() =>
        import("../automations/actions/ActionEditor.vue"),
    ),
};

const emit = defineEmits<{
    (e: 'open', component_name: string): void,
    (e: 'close'): void,
}>()

const componentCache = ref<KeyyValuePair<OpenPanelEvent>>({});
const presentationQueue = ref<Array<string>>([]);

const props = defineProps({
    item: {
        type: Object as PropType<OpenPanelEvent>,
        required: true
    },
});

watch(
    () => props.item,
    () => {
        openComponent(props.item)
    }, { immediate: true }
)

function openComponent(event: OpenPanelEvent): void {

    presentationQueue.value.push(event.name);

    console.log("OPEN:" + event.name, " owner: " + event.owner, " Size ", presentationQueue.value.length);

    if (componentCache.value[event.name] != undefined) {
        componentCache.value[event.name].args = event.args;

    } else {
        componentCache.value[event.name] = event;
    }
}

const currentComponent = computed(() => {
    const lastValue = presentationQueue.value.at(-1)
    return lastValue === undefined ? '' : lastValue;
});


function closeComponent(name: string): void {

    if (componentCache.value[name] === undefined) {
        console.log("CLOSE ", name, " NOT FOUND")
        return
    }

    // TODO: cleanup componentCache ?
    presentationQueue.value.pop();
    const event = componentCache.value[name];
   
    console.log("CLOSE:" + event.name, " owner: " + event.owner, " Size ", presentationQueue.value.length);
    if (presentationQueue.value.length == 0) {
        emit('close');
    }
}

//// ######################
const localBus = mitt<Events>();
provide(LocalEventBus, localBus);

const cleanup = useMyEvents({
    openPanel(e: OpenPanelEvent) {
        openComponent(e);
    },
    closePanel(name: string) {
        closeComponent(name)
    },
})


function useMyEvents(handlers: EventHandlers<Events>) {
    const keys = Object.keys(handlers) as Array<keyof Events>;
    for (const key of keys) {
        localBus.on(key, handlers[key] as never);
    }

    const cleanup = () => {
        for (const key of keys) {
            localBus.off(key, handlers[key] as never);
        }
    };
    onUnmounted(cleanup);
    return cleanup;
}

onMounted(() => {
});

onUnmounted(() => {
    console.log("Panel unmounted - deregister eventBus messages")
    cleanup();
    presentationQueue.value = [];
    componentCache.value = {};
});



</script>

<template>
    <!-- PANEL : {{ currentComponent }} <br> -->

    <div class="row">
        <button type="button" class="btn-close" aria-label="Close" @click="closeComponent(currentComponent)"></button>
        <div class="col">
            <!-- @delete="removeItem"
        @save="saveItem" @open="openComponent" @close="closeComponent"  -->
            <component :is="componentMap[currentComponent]" v-bind="componentCache[currentComponent].args"
                v-on="componentCache[currentComponent].events" />
        </div>
    </div>

</template>
