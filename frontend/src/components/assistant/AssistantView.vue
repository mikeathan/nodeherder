<script setup lang="ts">
  /*
   * Assistant screen. Chat behaviour stays in `useAssistant`; this component owns layout
   * only, reusing the shell's page header, buttons, dialog and empty states instead of a
   * second design language.
   */
  import { ref } from 'vue';
  import ChatInput from '@/components/assistant/chat/ChatInput.vue';
  import AssistantMessageList from '@/components/assistant/chat/ChatMessageList.vue';
  import HistoryList from '@/components/assistant/history/HistoryList.vue';
  import AssistantEmptyState from '@/components/assistant/states/EmptyState.vue';
  import AssistantNotConfigured from '@/components/assistant/states/NotConfiguredState.vue';
  import UiPageHeader from '@/components/ui/UiPageHeader.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiDialog from '@/components/ui/UiDialog.vue';
  import { useAssistant } from '@/composables/useAssistant';

  const {
    activeConversation,
    conversations,
    inputText,
    isLoading,
    isLoadingHistory,
    setChatContainer,
    sendMessage,
    startNewConversation,
    selectConversation,
    deleteConversation,
    isConfigured,
    notConfiguredReason,
  } = useAssistant();

  const historyOpen = ref(false);

  function select(id: string) {
    selectConversation(id);
    historyOpen.value = false;
  }

  function startNew() {
    startNewConversation();
    historyOpen.value = false;
  }

  function handleRerun(newText: string) {
    inputText.value = newText;
    sendMessage();
  }
</script>

<template>
  <div class="nh-assistant">
    <UiPageHeader title="Assistant" subtitle="Ask about your home. Device actions still require confirmation.">
      <template #actions>
        <UiButton icon="history" label="Conversations" @click="historyOpen = true">Conversations</UiButton>
        <UiButton variant="primary" icon="add" @click="startNew">New chat</UiButton>
      </template>
    </UiPageHeader>

    <div v-if="!isConfigured" class="nh-assistant-state">
      <AssistantNotConfigured :reason="notConfiguredReason" />
    </div>

    <div v-else class="nh-assistant-body">
      <aside class="nh-assistant-rail" aria-label="Conversations">
        <div class="nh-assistant-rail-head">
          <h2>Conversations</h2>
          <span>{{ conversations.length }}</span>
        </div>
        <HistoryList
          :conversations="conversations"
          :activeId="activeConversation?.id"
          @select="select"
          @delete="deleteConversation" />
      </aside>

      <section class="nh-assistant-main" aria-label="Chat">
        <div v-if="isLoadingHistory" class="nh-assistant-loading" aria-busy="true">
          <span class="nh-skel" />
          <span class="nh-skel" />
          <span class="nh-skel" />
        </div>

        <AssistantEmptyState
          v-else-if="activeConversation.messages.length === 0"
          v-model="inputText"
          :isLoading="isLoading"
          @send="sendMessage" />

        <template v-else>
          <div :ref="setChatContainer" class="nh-assistant-messages">
            <AssistantMessageList
              :messages="activeConversation.messages"
              :isLoading="isLoading"
              @rerun="handleRerun" />
          </div>
          <div class="nh-assistant-compose">
            <ChatInput v-model="inputText" :isLoading="isLoading" @send="sendMessage" />
          </div>
        </template>
      </section>
    </div>

    <UiDialog :open="historyOpen" title="Conversations" wide fullscreen-mobile @close="historyOpen = false">
      <HistoryList
        :conversations="conversations"
        :activeId="activeConversation?.id"
        @select="select"
        @delete="deleteConversation" />
      <template #footer>
        <span class="nh-muted">{{ conversations.length }} conversations</span>
        <UiButton icon="add" @click="startNew">New chat</UiButton>
      </template>
    </UiDialog>
  </div>
</template>

<style scoped>
  .nh-assistant {
    display: flex;
    flex-direction: column;
    min-height: calc(100dvh - var(--nh-header-h) - 3rem);
  }
  .nh-assistant-body {
    flex: 1;
    min-height: 0;
    display: grid;
    grid-template-columns: 16rem minmax(0, 1fr);
    gap: var(--nh-space);
  }
  .nh-assistant-rail {
    border: var(--nh-border-w) solid var(--nh-border);
    border-radius: var(--nh-radius-l);
    background: var(--nh-surface);
    padding: 0.6rem;
    min-width: 0;
    max-height: calc(100dvh - 11rem);
    overflow: auto;
  }
  .nh-assistant-rail-head {
    display: flex;
    align-items: baseline;
    gap: 0.4rem;
    padding: 0.25rem 0.4rem 0.5rem;
  }
  .nh-assistant-rail-head h2 {
    font-size: 0.8rem;
    font-weight: 700;
  }
  .nh-assistant-rail-head span {
    color: var(--nh-text-muted);
    font-size: 0.7rem;
  }
  .nh-assistant-main {
    display: flex;
    flex-direction: column;
    min-height: 0;
    border: var(--nh-border-w) solid var(--nh-border);
    border-radius: var(--nh-radius-l);
    background: var(--nh-surface);
    overflow: hidden;
  }
  .nh-assistant-messages {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    padding: var(--nh-space);
    display: flex;
    flex-direction: column;
    align-items: center;
  }
  .nh-assistant-compose {
    flex: none;
    border-top: var(--nh-border-w) solid var(--nh-border);
    padding: 0.7rem var(--nh-space) calc(0.7rem + env(safe-area-inset-bottom));
    background: var(--nh-surface);
  }
  .nh-assistant-state {
    display: grid;
    place-items: center;
    min-height: 18rem;
    border: var(--nh-border-w) solid var(--nh-border);
    border-radius: var(--nh-radius-l);
    background: var(--nh-surface);
  }
  .nh-assistant-loading {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    padding: var(--nh-space);
  }
  @media (max-width: 900px) {
    .nh-assistant-body {
      grid-template-columns: minmax(0, 1fr);
    }
    .nh-assistant-rail {
      display: none;
    }
    .nh-assistant-main {
      border-radius: var(--nh-radius-m);
    }
  }
</style>
