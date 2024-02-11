<script setup>
import { ref, onMounted, reactive } from 'vue';
const props = defineProps(['customClass']);
let tabContainer = ref(null);
let tabs = ref(null);
let activeTabIndex = ref(0);

onMounted(() => {
    tabs.value = [...tabContainer.value.querySelectorAll('.tab')];
    for (let x of tabs.value) {
        if (x.classList.contains('active')) {
            activeTabIndex = tabs.value.indexOf(x);
        }
    }
})
const changeTab = (index) => {

    // for (let x of [...tabs.value, ...tabHeaders.value]) {
    //     x.classList.remove('active')
    // }
    // tabs.value[activeTabIndex].classList.add('active')
    // tabHeaders.value[activeTabIndex].classList.add('active')

    activeTabIndex = index;
    for (let x of tabs.value) {
        x.classList.remove('active')
    }
    tabs.value[activeTabIndex].classList.add('active')
}

</script>

<template>
    <div id="tabs-container" :class="customClass" ref="tabContainer">

        <div id="tab-headers">
            <ul class="nav nav-tabs ">
                <li v-for=" (tab, index) in tabs" :key="index" class="nav-item" data-bs-toggle="tab" role="presentation">
                    <!-- @click.stop.prevent="setActive(tab)" -->
                    <a href="#" class="nav-link" rrole="tab" data-toggle="tab"
                        :class="activeTabIndex == index ? 'active' : ''" @click="changeTab(index)">{{ tab.title }}</a>
                </li>
            </ul>
            <ul id="active-tab" class="panel">
                <slot></slot>
            </ul>
        </div>

    </div>

    <!-- <div id="tabs-container" class="nav nav-tabs customClass" ref="tabContainer">
        <div id="tab-headers">
            <ul>
                <li v-for="(tab, index) in tabs" :key="index" :class="activeTabIndex == index ? 'active' : ''"
                    @click="changeTab(index)" ref="tabHeaders">{{ tab.title }}</li>
            </ul>
        </div>
        <li class="nav-item dropdown">
            <a class="nav-link dropdown-toggle" data-bs-toggle="dropdown" href="#" role="button"
                aria-expanded="false">Dropdown</a>
            <ul class="dropdown-menu">
                <slot></slot>
            </ul>
        </li>
    </div> -->
    <!-- <div id="tabs-container" :class="customClass" ref="tabContainer">
        <div id="tab-headers">
            <ul>
                <li v-for="(tab, index) in tabs" :key="index" :class="activeTabIndex == index ? 'active' : ''"
                    @click="changeTab(index)" ref="tabHeaders">{{ tab.title }}</li>
            </ul>
        </div>
        <div id="active-tab">
            <slot></slot>
        </div>
    </div> -->
</template>

<style></style>

