<script setup lang="tsx">
import { onBeforeMount, ref, h, resolveComponent, computed, PropType } from 'vue';
import { store } from '../../store/index';
import Status from './components/controls/Status.vue';
import Notifications from './components/hub/alerts/Notifications.vue';
import { useRouter } from 'vue-router';
import TimerButton from './components/controls/TimerButton.vue';
import { Button } from 'primevue';
import { onMounted } from 'vue';
type MenuItem = {
    to?: string;
    label: string;
    icon: string;
    command?: () => void;
    custom?: boolean;
    template?: () => void;
}


const props = defineProps({
    items: {
        type: Object as PropType<MenuItem[]>,
        default: [],
        required: true,
    }
})


const isMenuVisible = ref(true);
const isMobileView = ref(false);

const toggleMenu = () => {
    isMenuVisible.value = !isMenuVisible.value;
    console.log('Menu visibility:', isMenuVisible.value);
};
function checkMobileView() {
    isMobileView.value = window.innerWidth <= 768;
}

const renderMenuItem = (item: MenuItem) => {
    if (!item.custom) {
        return (
            <li class="p-menuitem">
                <a class="p-menuitem-link" href="javascript:void(0)" onClick={item.command}>
                    {item.icon ? <span class={`p-menuitem-icon ${item.icon}`}></span> : null}
                    <span class="p-menuitem-text">{item.label}</span>
                </a>
            </li>
        );
    } else if (item.template) {
        return (
            <li class="p-menuitem">
                <div>{item.template()}</div>
            </li>
        );
    }
    return null;
};


onMounted(() => {
    checkMobileView();
    window.addEventListener("resize", checkMobileView);
});

onBeforeMount(() => {
    window.removeEventListener("resize", checkMobileView);
});

</script>
<style scoped>
.p-menubar {
    background: var(--p-menubar-background);
    color: var(--primary-text-color, white);
    border: 1px solid var(--p-menubar-border-color);
}

.p-menubar-root-list {
    display: flex;
    list-style: none;
    margin: 0;
    padding: 0;
    background: var(--p-menubar-background);
    color: var(--primary-text-color, white);
}

.p-menuitem-link {
    display: flex;
    align-items: center;
    padding: 0.5rem 1rem;
    text-decoration: none;
    color: inherit;
    cursor: pointer;
}

.p-menuitem-link:hover {
    background-color: var(--surface-hover);
    color: var(--primary-color);
}

.p-menuitem-icon {
    margin-right: 0.5rem;
}

@media (max-width: 768px) {

    /* Initially hide Menubar in mobile */
    .p-menubar {
        display: none;
    }

    /* Ensure mobile view stacks menu items vertically */
    .p-menubar-root-list {
        flex-direction: column;
    }

    .p-menuitem {
        width: 100%;
    }
}

.hamburger {
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--p-menubar-background);
    color: var(--primary-text-color, white);
    padding: 0.5rem;
    cursor: pointer;
}

.hamburger i {
    font-size: 1.5rem;
}


.mobile-header {
    display: flex;
    justify-content: flex-end;
    background: var(--p-menubar-background);
    color: var(--primary-text-color, white);
    border: 1px solid var(--p-menubar-border-color);
    border-radius: var(--p-menubar-border-radius);
    padding: 0.5rem;
}
</style>

<template>

    <Menubar v-if="!isMobileView">
        <template #start>
            <ul class="p-menubar-root-list">
                <li v-for="(item, index) in props.items" :key="index" class="p-menuitem">
                    <component :is="renderMenuItem(item)" />

                    <!-- <template v-if="!item.custom">
                        <a class="p-menuitem-link" href="javascript:void(0)" @click="item.command && item.command()">
                            <span v-if="item.icon" :class="['p-menuitem-icon', item.icon]"></span>
                            <span class="p-menuitem-text">{{ item.label }}</span>
                        </a>
                    </template>
<template v-else>
                        <component v-if="item.template" :is="item.template()" />
                    </template> -->
                </li>
            </ul>
        </template>
    </Menubar>

    <!-- Mobile View -->
    <div v-else>
        <div class="mobile-header">
            <div class="hamburger" @click="toggleMenu">
                <i class="pi pi-bars"></i>
            </div>
        </div>
        <transition name="menu-slide">


            <ul v-if="isMenuVisible" class="p-menubar-root-list">

                <li v-for="(item, index) in props.items" :key="index" class="p-menuitem">
                    <component :is="renderMenuItem(item)" />

                    <!-- <template v-if="!item.custom">
                        <a class="p-menuitem-link" href="javascript:void(0)" @click="item.command && item.command()">
                            <span v-if="item.icon" :class="['p-menuitem-icon', item.icon]"></span>
                            <span class="p-menuitem-text">{{ item.label }}</span>
                        </a>
                    </template>
                    <template v-else>
                        <component v-if="item.template" :is="item.template()" />
                    </template> -->
                </li>

            </ul>
        </transition>
    </div>
</template>
