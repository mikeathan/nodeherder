<script setup lang="ts">
  /*
   * Button or link styled as a button (.nh-btn in base.css). With no default slot it renders
   * an icon-only button and `label` becomes its accessible name and tooltip.
   */
  import { computed, useSlots } from 'vue';
  import type { RouteLocationRaw } from 'vue-router';
  import UiIcon from './UiIcon.vue';
  import { IconName } from './icons';

  const props = withDefaults(
    defineProps<{
      variant?: 'default' | 'primary' | 'ghost' | 'danger';
      size?: 'md' | 'sm';
      icon?: IconName;
      label?: string;
      to?: RouteLocationRaw;
      type?: 'button' | 'submit';
      disabled?: boolean;
      loading?: boolean;
      block?: boolean;
      pressed?: boolean;
    }>(),
    { variant: 'default', size: 'md', type: 'button', pressed: undefined }
  );

  const slots = useSlots();
  const iconOnly = computed(() => !slots.default);
  const classes = computed(() => [
    'nh-btn',
    props.variant !== 'default' && `is-${props.variant}`,
    props.size === 'sm' && 'is-sm',
    iconOnly.value && 'is-icon',
    props.block && 'is-block',
  ]);
</script>

<template>
  <RouterLink v-if="to && !disabled" :to="to" :class="classes" :aria-label="iconOnly ? label : undefined" :title="iconOnly ? label : undefined">
    <UiIcon v-if="icon" :name="icon" />
    <slot />
  </RouterLink>
  <button
    v-else
    :type="type"
    :class="classes"
    :disabled="disabled || loading"
    :aria-busy="loading || undefined"
    :aria-pressed="pressed"
    :aria-label="iconOnly ? label : undefined"
    :title="iconOnly ? label : undefined">
    <span v-if="loading" class="nh-spin" aria-hidden="true" />
    <UiIcon v-else-if="icon" :name="icon" />
    <slot />
  </button>
</template>
