<script setup lang="ts">
  /*
   * Modal dialog: labelled, Escape and backdrop close it, focus moves inside on open, Tab is
   * kept inside, and focus returns to the opener on close. A bottom sheet on phones.
   */
  import { nextTick, onBeforeUnmount, ref, useId, watch } from 'vue';
  import UiButton from './UiButton.vue';

  const props = defineProps<{ open: boolean; title: string; wide?: boolean; dismissible?: boolean }>();
  const emit = defineEmits<{ (e: 'close'): void }>();

  const titleId = useId();
  const panel = ref<HTMLElement | null>(null);
  let opener: HTMLElement | null = null;

  const FOCUSABLE = 'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';
  const focusables = () => Array.from(panel.value?.querySelectorAll<HTMLElement>(FOCUSABLE) ?? []);

  function close() {
    if (props.dismissible !== false) emit('close');
  }

  function onKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      e.stopPropagation();
      close();
      return;
    }
    if (e.key !== 'Tab') return;
    const items = focusables();
    if (!items.length) return;
    const first = items[0];
    const last = items[items.length - 1];
    if (e.shiftKey && document.activeElement === first) {
      e.preventDefault();
      last.focus();
    } else if (!e.shiftKey && document.activeElement === last) {
      e.preventDefault();
      first.focus();
    }
  }

  watch(
    () => props.open,
    async (open) => {
      if (open) {
        opener = document.activeElement as HTMLElement | null;
        await nextTick();
        const body = panel.value?.querySelector<HTMLElement>('[autofocus], .nh-dialog-body ' + FOCUSABLE);
        (body ?? focusables()[0] ?? panel.value)?.focus();
      } else {
        opener?.focus?.();
        opener = null;
      }
    },
    { immediate: true }
  );
  onBeforeUnmount(() => opener?.focus?.());
</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="nh-dialog-root" @keydown="onKeydown">
      <div class="nh-dialog-backdrop" @click="close" />
      <div ref="panel" class="nh-dialog" :class="{ 'is-wide': wide }" role="dialog" aria-modal="true" :aria-labelledby="titleId" tabindex="-1">
        <header class="nh-dialog-head">
          <h2 :id="titleId">{{ title }}</h2>
          <UiButton v-if="dismissible !== false" variant="ghost" icon="close" label="Close" @click="close" />
        </header>
        <div class="nh-dialog-body"><slot /></div>
        <footer v-if="$slots.footer" class="nh-dialog-foot"><slot name="footer" /></footer>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
  .nh-dialog-backdrop {
    position: fixed;
    inset: 0;
    background: rgb(0 0 0 / 0.45);
    z-index: 1100;
    animation: nh-fade var(--nh-motion);
  }
  .nh-dialog {
    position: fixed;
    left: 50%;
    top: 50%;
    transform: translate(-50%, -50%);
    z-index: 1101;
    width: min(30rem, calc(100vw - 2rem));
    max-height: calc(100dvh - 2rem);
    display: flex;
    flex-direction: column;
    background: var(--nh-surface);
    color: var(--nh-text);
    border: var(--nh-border-w) solid var(--nh-border-strong);
    border-radius: var(--nh-radius-l);
    box-shadow: var(--nh-shadow-2);
    outline: none;
  }
  .nh-dialog.is-wide {
    width: min(44rem, calc(100vw - 2rem));
  }
  .nh-dialog-head {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    padding: 0.6rem 0.6rem 0.6rem var(--nh-space);
    border-bottom: var(--nh-border-w) solid var(--nh-border);
  }
  .nh-dialog-head h2 {
    font-size: 1.05rem;
    flex: 1;
  }
  .nh-dialog-body {
    padding: var(--nh-space);
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    overflow: auto;
  }
  .nh-dialog-body :deep(p) {
    margin: 0;
  }
  .nh-dialog-foot {
    display: flex;
    justify-content: flex-end;
    flex-wrap: wrap;
    gap: 0.5rem;
    padding: 0.75rem var(--nh-space);
    border-top: var(--nh-border-w) solid var(--nh-border);
  }
  @media (max-width: 640px) {
    .nh-dialog,
    .nh-dialog.is-wide {
      left: 0;
      right: 0;
      top: auto;
      bottom: 0;
      transform: none;
      width: 100%;
      max-height: 92dvh;
      border-radius: var(--nh-radius-l) var(--nh-radius-l) 0 0;
      padding-bottom: env(safe-area-inset-bottom);
    }
    .nh-dialog-foot > :deep(*) {
      flex: 1 1 auto;
    }
  }
</style>
