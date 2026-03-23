<script setup lang="ts">
  import { ref } from 'vue';
  import type { MessageAlignment } from '@/types/controls.type';
  import { copyToClipboard } from '@/utils/clipboard.utils';

  const props = defineProps<{
    alignment: MessageAlignment;
    textToCopy: string;
  }>();

  defineSlots<{
    avatar(props: {}): any;
    content(props: {}): any;
    actions(props: {}): any;
  }>();

  const copySuccess = ref(false);

  const copyText = async () => {
    try {
      await copyToClipboard(props.textToCopy);
      copySuccess.value = true;
      setTimeout(() => {
        copySuccess.value = false;
      }, 2000);
    } catch (err) {
    }
  };
</script>

<template>
  <div
    class="flex w-full bubble-wrapper"
    :class="alignment === 'right' ? 'justify-content-end' : 'justify-content-start'"
    tabindex="0"
  >
    <div class="flex gap-4 w-full" :class="alignment === 'right' ? 'justify-content-end' : 'align-items-start'">
      <!-- Avatar -->
      <div v-if="$slots.avatar" class="mt-1">
        <slot name="avatar"></slot>
      </div>

      <!-- Content Column -->
      <div
        class="flex flex-column gap-1"
        :class="alignment === 'left' ? 'flex-1' : ''"
        :style="[alignment === 'right' ? { maxWidth: '36rem' } : { width: '100%' }]">
        <slot name="content"></slot>

        <!-- Action Bar at Bottom  -->
        <div
          class="flex align-items-center justify-content-start transition-opacity action-bar"
          @click.stop>
          <slot name="actions"></slot>
          <Button
            :icon="copySuccess ? 'pi pi-check' : 'pi pi-copy'"
            :class="copySuccess ? 'text-green-500' : 'text-color-secondary'"
            text
            rounded
            class="p-button-sm p-0 flex align-items-center justify-content-center opacity-70 hover:opacity-100 transition-opacity"
            style="width: 28px; height: 28px"
            v-tooltip.bottom="'Copy text'"
            @click="copyText" />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.action-bar {
  opacity: 0;
  pointer-events: none; /* Prevent clicking when invisible */
}

/* Show actions on hover (Desktop) and when tapped/focused (Touch devices) */
.bubble-wrapper:hover .action-bar,
.bubble-wrapper:focus .action-bar,
.bubble-wrapper:focus-within .action-bar {
  opacity: 1;
  pointer-events: auto;
}
</style>
