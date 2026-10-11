<script setup lang="ts">
  /* Card with optional header (title, icon, tools) and footer. `flush` removes body padding. */
  import UiIcon from './UiIcon.vue';
  import { IconName } from './icons';

  defineProps<{ title?: string; icon?: IconName; flush?: boolean; headingLevel?: 2 | 3 }>();
</script>

<template>
  <section class="nh-card">
    <header v-if="title || $slots.tools" class="nh-card-head">
      <component :is="`h${headingLevel ?? 2}`" v-if="title" class="nh-card-title">
        <UiIcon v-if="icon" :name="icon" />
        {{ title }}
      </component>
      <div v-if="$slots.tools" class="nh-card-tools"><slot name="tools" /></div>
    </header>
    <div :class="{ 'nh-card-body': !flush }"><slot /></div>
    <footer v-if="$slots.foot" class="nh-card-foot"><slot name="foot" /></footer>
  </section>
</template>
