<script setup lang="ts">
import { ref, watch, PropType, defineAsyncComponent, computed, onMounted, inject, onUnmounted } from "vue";
import { Emitter } from 'mitt'
import { EventActions, Events, OpenPanelEvent } from "@/types/events.type";




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
    "Action": defineAsyncComponent(() =>
        import("../automations/actions/Action.vue"),
    ),
};

const emit = defineEmits<{
    //(e: 'save', item: any): void,
    (e: 'delete', item: any): void,
    (e: 'open', component_name: string): void,
    (e: 'close'): void,
}>()

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
    }

});

function openComponent(component_name: string, args: any, events: EventActions): void {

    componentsCache.value.push({ name: component_name, args: args, events: events })
    console.log("open component: ", currentComponent.value.name, " history:  size ", componentsCache.value.length)
}

function closeComponent(): void {
    const prevSz = componentsCache.value.length;
    const prevName = currentComponent.value.name
    const lastComponent = componentsCache.value.pop();
    if (lastComponent != undefined) {
        lastComponent.events = {};
    }

    console.log("close component: ", prevName, " history previous size: ", prevSz, " size", componentsCache.value.length);

    // PROBLEM HERE
    if (componentsCache.value.length == 0) {
        console.log("no more components. emit panel close to parent");

        emit('close');
    }
}

const currentComponent = computed(() => {// problem

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
        openComponent(props.component_name, props.component_props, props.component_events)
    }, { immediate: true }
)

let eventBus = inject('emitter') as Emitter<Events>;
onMounted(() => {

    console.log("Panel mounted - register eventBus messages")

    eventBus.on('openPanel', (e: OpenPanelEvent) => {
        console.log("openpanel received ", e.name)
        openComponent(e.name, e.args, e.events);
    });

    eventBus.on('closePanel', (e: string) => {
        console.log('closepanel event received from:', e);
        closeComponent()
    });
});

onUnmounted(() => {
    console.log("Panel unmounted - deregister eventBus messages")

    eventBus.off('openPanel', (e: OpenPanelEvent) => {

    });

    eventBus.off('closePanel', (e: string) => {

    });
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
                v-on="currentComponent.events" />
        </div>
    </div>

</template>


<!-- TODO

import mitt from 'mitt'

const emitter = mitt()

// listen to an event
emitter.on('foo', e => console.log('foo', e) )

// listen to all events
emitter.on('*', (type, e) => console.log(type, e) )

// fire an event
emitter.emit('foo', { a: 'b' })

// clearing all events
emitter.all.clear()

// working with handler references:
function onFoo() {}
emitter.on('foo', onFoo)   // listen
emitter.off('foo', onFoo)  // unlisten

------------------------
import mitt from 'mitt';

type Events = {
  foo: string;
  bar?: number;
};

const emitter = mitt<Events>(); // inferred as Emitter<Events>

emitter.on('foo', (e) => {}); // 'e' has inferred type 'string'

emitter.emit('foo', 42); // Error: Argument of type 'number' is not assignable to parameter of -->