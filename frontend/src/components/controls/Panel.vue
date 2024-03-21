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
    component_name: {
        type: String,
        required: true
    },
    component_props: {
        type: Object,
        required: true
    },
    component_events: {
        type: Object,
        required: true
    },
    overrideEvents: {
        type: Boolean,
        required: true
    }

});

function openComponent(component_name: string, args: any, events: EventActions, overrideEvents: boolean): void {
    const prevSz = componentsCache.value.length;

    if (componentEventsCache.value[component_name] != undefined || overrideEvents) {
        componentEventsCache.value[component_name] = events;
    }
    componentsCache.value.push({ name: component_name, args: args, events: events, overrideEvents: overrideEvents })
    console.log("OpenPanel received from:", component_name, " CurrentComponent: ", currentComponent.value.name, " Cache was:", prevSz, " is: ", componentsCache.value.length)
}

function closeComponent(): void {
    const prevSz = componentsCache.value.length;
    const prevName = currentComponent.value.name
    const lastComponent = componentsCache.value.pop();
    // if (lastComponent != undefined) {
    //     lastComponent.events = {};
    // }

    console.log("ClosePane:l " + prevName + " CurrentComponent: ", currentComponent.value.name, " Cache was:", prevSz, " is: ", componentsCache.value.length)

    // PROBLEM HERE
    if (componentsCache.value.length == 0) {
        console.log("COMPONENT IS EMPTY. emit panel close to parent");

        emit('close');
    }
}
const currentComponentEvents = computed(() => {

    const lastValue = componentsCache.value.at(-1)
    const name = lastValue === undefined ? '' : lastValue.name;
    return componentEventsCache.value[name];
});

const currentComponent = computed(() => {

    const lastValue = componentsCache.value.at(-1)
    return lastValue === undefined ? { name: '', args: '', events: {} } : lastValue as OpenPanelEvent;
});

function saveItem(item: any): void {
    /// emit('save', item);
}

function removeItem(item: any): void {
    // emit('delete', item);
}

watch(
    () => props.component_name,
    () => {
        openComponent(props.component_name, props.component_props, props.component_events, props.overrideEvents)
    }, { immediate: true }
)

//let eventBus = inject('emitter') as Emitter<Events>;


//// ######################
const localBus = mitt<Events>();
provide(LocalEventBus, localBus);

const cleanup = useMyEvents({
    openPanel(e: OpenPanelEvent) {
        openComponent(e.name, e.args, e.events, e.overrideEvents);
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
            <component :is="componentMap[currentComponent.name]" v-bind="currentComponent.args"
                v-on="currentComponentEvents" />
        </div>
    </div>

</template>
