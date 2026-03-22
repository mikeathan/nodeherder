<script setup lang="ts">
  import { ref, computed } from 'vue';
  import ChatInput from '@/components/assistant/core/ChatInput.vue';
  import AssistantMessageList from '@/components/assistant/core/AssistantMessageList.vue';
  import AssistantHistoryList from '@/components/assistant/core/AssistantHistoryList.vue';
  import AssistantEmptyState from '@/components/assistant/state/AssistantEmptyState.vue';
  import AssistantNotConfigured from '@/components/assistant/state/AssistantNotConfigured.vue';
  import { useAssistant } from '@/composables/useAssistant';

  const {
    activeConversation,
    conversations,
    inputText,
    isLoading,
    isLoadingHistory,
    chatContainer,
    sendMessage,
    startNewConversation,
    selectConversation,
    deleteConversation,
    isConfigured,
  } = useAssistant();

  const showMobileHistory = ref(false);
  const showDesktopHistory = ref(false); // Default to false as requested

  const isHistoryOpen = computed(() => {
    // Check which one is active depending on what view we assume
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
    selectConversation(id);
    showMobileHistory.value = false;
  };
</script>

<template>
  <div class="assistant-container flex h-full w-full surface-ground relative">
    
    <!-- 1. The New Inner Sidebar (History Panel) -->
    <!-- Hidden on small screens (hidden md:flex) -->
    <div v-if="showDesktopHistory" class="hidden md:flex flex-column w-16rem surface-section border-right-1 surface-border h-full transition-all transition-duration-300">
       <div class="p-3">
         <Button icon="pi pi-plus" label="New Chat" outlined class="w-full" @click="startNewConversation" />
       </div>
       <div class="flex-1 overflow-y-auto">
         <AssistantHistoryList 
           :conversations="conversations" 
           :activeId="activeConversation?.id"
           @select="handleSelect" 
           @delete="deleteConversation" 
         />
       </div>
    </div>

    <!-- Mobile Sidebar Drawer Overlay -->
    <Sidebar v-model:visible="showMobileHistory" class="w-16rem p-sidebar-sm">
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
            @click="startNewConversation(); showMobileHistory = false;" 
          />
        </div>
        <div class="flex-1 overflow-y-auto mt-2">
          <AssistantHistoryList 
            :conversations="conversations" 
            :activeId="activeConversation?.id"
            @select="handleSelect" 
            @delete="deleteConversation" 
          />
        </div>
      </div>
    </Sidebar>

    <!-- 2. The Existing Chat Area -->
    <div class="flex-1 flex flex-column h-full relative min-w-0">
      
      <!-- History toggle button (visible on all screens) -->
      <div class="absolute top-0 left-0 p-3 z-5">
        <Button 
          :icon="toggleIcon" 
          text 
          rounded 
          class="p-button-secondary" 
          :class="{ 'surface-hover': showDesktopHistory }"
          @click="toggleHistory" 
          aria-label="Toggle Chat History" 
          v-tooltip.bottom="'Toggle Chat History'"
        />
      </div>

      <div class="flex flex-column h-full pt-4">
        <AssistantNotConfigured v-if="!isConfigured" />

        <template v-else>
          <div v-if="isLoadingHistory" class="flex-1 flex align-items-center justify-content-center">
             <i class="pi pi-spin pi-spinner text-4xl text-color-secondary"></i>
          </div>

          <template v-else>
            <AssistantEmptyState
              v-if="activeConversation.messages.length === 0"
              v-model="inputText"
              :isLoading="isLoading"
              @send="sendMessage" />

            <template v-else>
              <div
                ref="chatContainer"
                class="flex-1 overflow-y-auto px-4 py-4 flex flex-column align-items-center w-full min-h-0 messages-container">
                <AssistantMessageList :messages="activeConversation.messages" :isLoading="isLoading" />
              </div>

              <div class="w-full flex justify-content-center pb-6 px-4 surface-ground pt-3 relative z-2 flex-shrink-0">
                <div class="w-full">
                  <ChatInput v-model="inputText" :isLoading="isLoading" @send="sendMessage" />
                </div>
              </div>
            </template>
          </template>
        </template>
      </div>

    </div>
  </div>
</template>

<style scoped>
  .assistant-container {
    height: 100%;
  }
  .messages-container {
    scroll-behavior: smooth;
  }
</style>
