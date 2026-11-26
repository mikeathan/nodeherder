<script setup lang="ts">
  /* Conversation list with app-token rows; delete stays visible on touch and on focus. */
  import type { ConversationSummary } from '@/types/assistant.type';
  import UiButton from '@/components/ui/UiButton.vue';

  defineProps<{
    conversations: ConversationSummary[];
    activeId?: string;
  }>();

  const emit = defineEmits<{
    (e: 'select', id: string): void;
    (e: 'delete', id: string): void;
  }>();

  function formatDate(isoString?: string): string {
    if (!isoString) return '';
    return new Date(isoString).toLocaleDateString([], { month: 'short', day: 'numeric' });
  }
</script>

<template>
  <ul class="nh-conv-list">
    <li v-for="conversation in conversations" :key="conversation.id">
      <div class="nh-conv" :class="{ 'is-active': conversation.id === activeId }">
        <button type="button" class="nh-conv-main" @click="emit('select', conversation.id)">
          <span>{{ conversation.title || 'New conversation' }}</span>
          <small>{{ formatDate(conversation.created_at) }}</small>
        </button>
        <UiButton
          size="sm"
          variant="ghost"
          icon="delete"
          class="is-danger nh-conv-delete"
          :label="`Delete ${conversation.title || 'conversation'}`"
          @click.stop="emit('delete', conversation.id)" />
      </div>
    </li>
    <li v-if="!conversations.length" class="nh-conv-empty">No recent conversations</li>
  </ul>
</template>

<style scoped>
  .nh-conv-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
  }
  .nh-conv {
    display: flex;
    align-items: center;
    gap: 0.2rem;
    border-radius: var(--nh-radius-m);
    padding: 0.1rem;
  }
  .nh-conv:hover,
  .nh-conv.is-active {
    background: var(--nh-surface-2);
  }
  .nh-conv.is-active {
    box-shadow: inset 3px 0 0 var(--nh-accent);
  }
  .nh-conv-main {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 0.1rem;
    border: 0;
    background: none;
    color: inherit;
    font: inherit;
    text-align: left;
    padding: 0.45rem 0.55rem;
    cursor: pointer;
  }
  .nh-conv-main span {
    font-size: 0.8rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 100%;
  }
  .nh-conv-main small {
    color: var(--nh-text-muted);
    font-size: 0.68rem;
  }
  .nh-conv-delete {
    opacity: 0;
    flex: none;
  }
  .nh-conv:hover .nh-conv-delete,
  .nh-conv:focus-within .nh-conv-delete {
    opacity: 1;
  }
  .nh-conv-empty {
    color: var(--nh-text-muted);
    font-size: 0.78rem;
    padding: 0.6rem;
    text-align: center;
  }
  @media (hover: none) {
    .nh-conv-delete {
      opacity: 1;
    }
  }
</style>
