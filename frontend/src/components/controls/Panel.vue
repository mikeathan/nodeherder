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
    //(e: 'save', item: any): void,
    //(e: 'delete', item: any): void,
    (e: 'open', component_name: string): void,
    (e: 'close'): void,
}>()

const componentEventsCache = ref<KeyyValuePair<EventActions>>({});
const componentsCache = ref<Array<OpenPanelEvent>>([]);
const props = defineProps({


    item: {
        type: Object as PropType<OpenPanelEvent>,
        required: true
    },


});

function openComponent(event: OpenPanelEvent): void {
    if (componentEventsCache.value[event.target] != undefined && event.events === null) {
        console.log("OpenPanel received: ", event.target, " exists in cache. [Dont override.]")

    } else {
        const prevSz = componentsCache.value.length;
        componentsCache.value.push(event)
        console.log("OpenPanel received: ", currentComponent.value.target, " Cache was:", prevSz, " is: ", componentsCache.value.length)
    }
}

function closeComponent(): void {
    const prevSz = componentsCache.value.length;
    const prevName = currentComponent.value.target


    // NEED LOGIC HERE
    // IF CLOSE IS NOT FOR SOURCE, THEN WE CANT REMOVE THE EVENT - NNED TOO EITHER CHANGE POSITION OR CACHE IT
    // IF CLOSE IS FOR SOURCE THEN WE ARE GOOD TO CLEAR
    // WE NEED TO KEEP THE EVENTS AND ITEMS FOR SOURCE EVENT 

    const lastComponent = componentsCache.value.pop();
    // if (lastComponent != undefined) {
    //     lastComponent.events = {};
    // }

    console.log("ClosePane:l " + prevName + " CurrentComponent: ", currentComponent.value.target, " Cache was:", prevSz, " is: ", componentsCache.value.length)

    // PROBLEM HERE
    if (componentsCache.value.length == 0) {
        console.log("COMPONENT IS EMPTY. emit panel close to parent");

        emit('close');
    }
}

const currentComponent = computed(() => {
    const lastValue = componentsCache.value.at(-1)
    return lastValue === undefined ? { source: '', target: '', args: '', events: {} } : lastValue as OpenPanelEvent;
});


watch(
    () => props.item,
    () => {
        openComponent(props.item)
    }, { immediate: true }
)


//// ######################
const localBus = mitt<Events>();
provide(LocalEventBus, localBus);

const cleanup = useMyEvents({
    openPanel(e: OpenPanelEvent) {
        openComponent(e);
    },
    closePanel(e: string) {
        console.log('closePanel event received from:', e);
        closeComponent()
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

//// ######################
onMounted(() => {


    // console.log("Panel mounted - register eventBus messages")

    // localBus.on('openPanel', (e: OpenPanelEvent) => {
    //     //console.log("OpenPanel received from:", e.name)
    //     openComponent(e.name, e.args, e.events);
    // });

    // localBus.on('closePanel', (e: string) => {
    //     console.log('OpenPanel event received from:', e);
    //     closeComponent()
    // });
});

onUnmounted(() => {
    console.log("Panel unmounted - deregister eventBus messages")
    cleanup();
});


// component that we pass in can raise event to be changed


</script>

<template>
    <!-- PANEL : {{ currentComponent }} <br> -->

    <div class="row">
        <button type="button" class="btn-close" aria-label="Close" @click="closeComponent"></button>
        <div class="col">
            <!-- @delete="removeItem"
        @save="saveItem" @open="openComponent" @close="closeComponent"  -->
            <component :is="componentMap[currentComponent.target]" v-bind="currentComponent.args"
                v-on="currentComponent.events" />
        </div>
    </div>

</template>
