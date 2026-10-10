<script setup lang="ts">
  /* One expose as a labelled row: kind icon, name, and its value or control. */
  import { computed } from 'vue';
  import { Device, Expose } from '@/types/device';
  import { exposeKind, exposeLabel, isBinaryOn } from '@/domain/exposes';
  import { exposeIcon } from '@/components/ui/icons';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import ExposeControl from './ExposeControl.vue';

  const props = defineProps<{ device: Device; expose: Expose; showDescription?: boolean }>();
  const label = computed(() => exposeLabel(props.expose.name));
  const icon = computed(() => exposeIcon(props.expose.name, isBinaryOn(props.expose)));
</script>

<template>
  <div class="nh-row" :class="`kind-${exposeKind(expose.name)}`">
    <UiIcon :path="icon" class="nh-row-ic" />
    <div class="nh-row-label">
      <span>{{ label }}</span>
      <small v-if="showDescription && expose.description">{{ expose.description }}</small>
    </div>
    <ExposeControl :device="device" :expose="expose" :label="label" />
  </div>
</template>

<style scoped>
  .nh-row {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 0.4rem 0.65rem;
    padding: 0.45rem var(--nh-space);
    min-height: calc(var(--nh-ctl-h) + 0.25rem);
  }
  .nh-row + .nh-row {
    border-top: var(--nh-border-w) solid var(--nh-border);
  }
  .nh-row-ic {
    color: var(--nh-kc, var(--nh-text-muted));
  }
  .nh-row-label {
    flex: 1 1 6rem;
    display: flex;
    flex-direction: column;
    min-width: 0;
    font-size: 0.9rem;
  }
  .nh-row-label span {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .nh-row-label small {
    color: var(--nh-text-muted);
  }
</style>
