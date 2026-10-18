<script setup lang="ts">
  /*
   * Modal dialog: labelled, Escape and backdrop close it, focus moves inside on open, Tab is
   * kept inside, and focus returns to the opener on close. A bottom sheet on phones.
   */
  import { nextTick, onBeforeUnmount, ref, useId, watch } from 'vue';
  import UiButton from './UiButton.vue';

  const props = withDefaults(
    defineProps<{ open: boolean; title: string; wide?: boolean; dismissible?: boolean; fullscreenMobile?: boolean }>(),
    { dismissible: true }
  );
  const emit = defineEmits<{ (e: 'close'): void }>();

  const titleId = useId();
  const panel = ref<HTMLElement | null>(null);
  let opener: HTMLElement | null = null;
  let previousOverflow: string | null = null;

  function unlockScroll() {
    document.removeEventListener('keydown', onDocumentKeydown, true);
    if (previousOverflow === null) return;
    document.body.style.overflow = previousOverflow;
    previousOverflow = null;
  }

  const FOCUSABLE =
    'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), summary, [tabindex]:not([tabindex="-1"])';
  const focusables = () =>
    Array.from(panel.value?.querySelectorAll<HTMLElement>(FOCUSABLE) ?? []).filter(
      (element) => element.getClientRects().length > 0
    );

  function close() {
    if (props.dismissible !== false) emit('close');
  }

  function onKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      e.stopPropagation();
      close();
      return;
    }
    if (e.key === 'Tab') trapFocus(e);
  }

  function trapFocus(event: KeyboardEvent) {
    const items = focusables();
    if (!items.length) return;
    event.preventDefault();
    const current = items.indexOf(document.activeElement as HTMLElement);
    const direction = event.shiftKey ? -1 : 1;
    const index =
      current < 0 ? (event.shiftKey ? items.length - 1 : 0) : (current + direction + items.length) % items.length;
    items[index].focus();
  }

  // Pending controls may lose focus when disabled; Escape must still dismiss the modal.
  function onDocumentKeydown(event: KeyboardEvent) {
    const dialogs = document.querySelectorAll('[role="dialog"][aria-modal="true"]');
    if (dialogs[dialogs.length - 1] !== panel.value) return;
    onKeydown(event);
  }

  watch(
    () => props.open,
    async (open) => {
      if (open) {
        previousOverflow = document.body.style.overflow;
        document.body.style.overflow = 'hidden';
        document.addEventListener('keydown', onDocumentKeydown, true);
        opener = document.activeElement as HTMLElement | null;
        await nextTick();
        const body = panel.value?.querySelector<HTMLElement>('[autofocus], .nh-dialog-body ' + FOCUSABLE);
        (body ?? focusables()[0] ?? panel.value)?.focus();
      } else {
        unlockScroll();
        opener?.focus?.();
        opener = null;
      }
    },
    { immediate: true }
  );
  onBeforeUnmount(() => {
    unlockScroll();
    opener?.focus?.();
  });
</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="nh-dialog-root">
      <div class="nh-dialog-backdrop" @click="close" />
      <div
        ref="panel"
        class="nh-dialog"
        :class="{ 'is-wide': wide, 'is-fullscreen-mobile': fullscreenMobile }"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="titleId"
        tabindex="-1">
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
    .nh-dialog.is-wide.is-fullscreen-mobile,
    .nh-dialog.is-fullscreen-mobile:not(.is-wide) {
      top: 0;
      height: 100dvh;
      max-height: 100dvh;
      border: 0;
      border-radius: 0;
      padding-top: env(safe-area-inset-top);
    }
    .is-fullscreen-mobile .nh-dialog-body {
      flex: 1;
      padding: 1rem 1.5rem;
    }
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
