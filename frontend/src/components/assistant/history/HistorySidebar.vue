<script setup lang="ts">
import { ref, computed } from 'vue';
import HistoryList from './HistoryList.vue';
import { ConversationSummary } from '@/types/assistant.type';

defineProps<{
  conversations: ConversationSummary[];
  activeId?: string;
}>();

const emit = defineEmits<{
  (e: 'new-chat'): void;
  (e: 'select', id: string): void;
  (e: 'delete', id: string): void;
}>();

const showMobileHistory = ref(false);
const showDesktopHistory = ref(false);

const isHistoryOpen = computed(() => {
  if (typeof window !== 'undefined' && window.innerWidth >= 768) {
    return showDesktopHistory.value;
  }
  return showMobileHistory.value;
});

const toggleIcon = computed(() => {
  return isHistoryOpen.value ? 'pi pi-angle-double-left' : 'pi pi-angle-double-right';
});

const toggleHistory = () => {
  if (window.innerWidth >= 768) {
    showDesktopHistory.value = !showDesktopHistory.value;
  } else {
    showMobileHistory.value = !showMobileHistory.value;
  }
};

const handleSelect = (id: string) => {
  emit('select', id);
  showMobileHistory.value = false;
};
</script>

<template>
  <div class="assistant-container flex h-full w-full surface-ground relative">
    <!-- 1. History side panel -->
    <div
      v-if="showDesktopHistory"
      class="hidden md:flex flex-column w-16rem surface-section border-right-1 surface-border h-full transition-all transition-duration-300">
      <div class="p-3">
        <Button icon="pi pi-plus" label="New Chat" outlined class="w-full" @click="emit('new-chat')" />
      </div>
      <div class="flex-1 overflow-y-auto">
        <HistoryList
          :conversations="conversations"
          :activeId="activeId"
          @select="handleSelect"
          @delete="emit('delete', $event)" />
      </div>
    </div>

    <!-- Mobile Sidebar Drawer Overlay -->
    <Sidebar :visible="showMobileHistory" @update:visible="showMobileHistory = $event" class="w-16rem p-sidebar-sm">
      <template #header>
        <span class="font-semibold text-lg text-color">Chat History</span>
      </template>
      <div class="flex flex-column h-full">
        <div class="pb-3 border-bottom-1 surface-border">
          <Button
            icon="pi pi-plus"
            label="New Chat"
            class="w-full"
            outlined
            @click="
              emit('new-chat');
              showMobileHistory = false;
            " />
        </div>
        <div class="flex-1 overflow-y-auto mt-2">
          <HistoryList
            :conversations="conversations"
            :activeId="activeId"
            @select="handleSelect"
            @delete="emit('delete', $event)" />
        </div>
      </div>
    </Sidebar>

    <!-- 2. Main Content Slot Area -->
    <div class="flex-1 flex flex-column h-full relative min-w-0">
      <!-- History toggle button -->
      <div class="absolute top-0 left-0 p-3 z-5">
        <Button
          :icon="toggleIcon"
          text
          rounded
          class="p-button-secondary"
          :class="{ 'surface-hover': showDesktopHistory }"
          @click="toggleHistory"
          aria-label="Toggle Chat History"
          v-tooltip.bottom="'Toggle Chat History'" />
      </div>

      <!-- Provide the slotted child content here -->
      <slot></slot>
    </div>
  </div>
</template>

<style scoped>
  .assistant-container {
    height: 100%;
  }
</style>
