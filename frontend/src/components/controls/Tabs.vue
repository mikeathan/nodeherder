<script setup>
import { ref, onMounted, reactive } from 'vue';
const props = defineProps(['customClass']);
let tabContainer = ref(null);
let tabHeaders = ref(null);
let tabs = ref(null);
let activeTabIndex = ref(0);

onMounted(() => {
    tabs.value = [...tabContainer.value.querySelectorAll('.tab')];
    console.log(tabs.value)
    for (let x of tabs.value) {
        if (x.classList.contains('active')) {
            activeTabIndex = tabs.value.indexOf(x);
        }
    }
})
const changeTab = (index) => {
    activeTabIndex = index;
    for (let x of [...tabs.value, ...tabHeaders.value]) {
        x.classList.remove('active')
    }
    tabs.value[activeTabIndex].classList.add('active')
    tabHeaders.value[activeTabIndex].classList.add('active')
}
</script>

<template>
    <div id="tabs-container" :class="customClass" ref="tabContainer">

        <div id="tab-headers">
            <ul class="nav nav-tabs ">
                <li v-for=" (tab, index) in tabs" :key="index" class="nav-item" data-bs-toggle="tab" role="presentation"
                    ref="tabHeaders">
                    <a href="#" class="nav-link">{{ tab.title }} </a>
                </li>
            </ul>
            <ul class="dropdown-menu">
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

<style>
/* #tab-headers ul {
    margin: 0;
    padding: 0;
    display: flex;
    border-bottom: 2px solid #ddd;
}

#tab-headers ul li {
    list-style: none;
    padding: 1rem 1.25rem;
    position: relative;
    cursor: pointer;
}

#tab-headers ul li.active {
    color: #008438;
    font-weight: bold;
}

#tab-headers ul li.active:after {
    content: '';
    position: absolute;
    bottom: -2px;
    left: 0;
    height: 2px;
    width: 100%;
    background: #008438;
}

#active-tab,
#tab-headers {
    width: 100%;
}

#active-tab {
    padding: 0.75rem;
} */
</style>



<!-- <ul class="nav nav-tabs">
    <li class="nav-item">
      <a class="nav-link active" aria-current="page" href="#">Active</a>
    </li>
    <li class="nav-item dropdown">
      <a class="nav-link dropdown-toggle" data-bs-toggle="dropdown" href="#" role="button" aria-expanded="false">Dropdown</a>
      <ul class="dropdown-menu">
        <li><a class="dropdown-item" href="#">Action</a></li>
        <li><a class="dropdown-item" href="#">Another action</a></li>
        <li><a class="dropdown-item" href="#">Something else here</a></li>
        <li><hr class="dropdown-divider"></li>
        <li><a class="dropdown-item" href="#">Separated link</a></li>
      </ul>
    </li>
    <li class="nav-item">
      <a class="nav-link" href="#">Link</a>
    </li>
    <li class="nav-item">
      <a class="nav-link disabled" href="#" tabindex="-1" aria-disabled="true">Disabled</a>
    </li>
  </ul> -->