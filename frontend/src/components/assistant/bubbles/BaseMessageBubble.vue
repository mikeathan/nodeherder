<script setup lang="ts">
  import { ref } from 'vue';
  import type { MessageAlignment } from '@/types/controls.type';

  const props = defineProps<{
    alignment: MessageAlignment;
    textToCopy: string;
  }>();

  const isHovering = ref(false);
  const copySuccess = ref(false);

  const copyText = async () => {
    try {
      await navigator.clipboard.writeText(props.textToCopy);
      copySuccess.value = true;
      setTimeout(() => {
        copySuccess.value = false;
      }, 2000);
    } catch (err) {
      console.error('Failed to copy text: ', err);
    }
  };
</script>

<template>
  <div
    class="flex w-full"
    :class="alignment === 'right' ? 'justify-content-end' : 'justify-content-start'"
    @mouseenter="isHovering = true"
    @mouseleave="isHovering = false">
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
          class="flex align-items-center justify-content-start transition-opacity"
          :class="isHovering ? 'opacity-100' : 'opacity-0'">
          <Button
            :icon="copySuccess ? 'pi pi-check' : 'pi pi-copy'"
            :class="copySuccess ? 'text-green-500' : 'text-color-secondary'"
            text
            rounded
            class="p-button-sm p-0 flex align-items-center justify-content-center opacity-70 hover:opacity-100 transition-opacity"
            style="width: 28px; height: 28px"
            @click="copyText" />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped></style>
