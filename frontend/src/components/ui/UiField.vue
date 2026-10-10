<script setup lang="ts">
  /* Labelled form field with help and error text wired to the control via aria attributes. */
  import UiIcon from './UiIcon.vue';

  defineProps<{ label: string; for: string; help?: string; error?: string | null }>();
</script>

<template>
  <div class="nh-field">
    <label :for="$props.for">{{ label }}</label>
    <slot />
    <p v-if="error" :id="`${$props.for}-err`" class="nh-err" role="alert"><UiIcon name="error" />{{ error }}</p>
    <p v-else-if="help" :id="`${$props.for}-help`" class="nh-help">{{ help }}</p>
  </div>
</template>
