<script setup>
import { ref, onMounted } from 'vue';
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
    activeTabIndex = index;
    for (let x of tabs.value) {
        x.classList.remove('active')
    }
    tabs.value[activeTabIndex].classList.add('active')
}

</script>

<template>
    <div id="tab" :class="customClass" ref="tabContainer">

        <ul class="nav nav-tabs">
            <li v-for=" (tab, index) in tabs" :key="index" class="nav-item" data-bs-toggle="tab" role="presentation">
                <a href="#" class="nav-link" role="tab" data-toggle="tab"
                    :class="activeTabIndex == index ? 'active' : ''" @click="changeTab(index)">{{ tab.title }}</a>
            </li>
        </ul>
        <ul id="active-tab" class="panel">
            <slot></slot>
        </ul>

    </div>
</template>
