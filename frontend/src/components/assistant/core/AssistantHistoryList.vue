<script setup lang="ts">
import { ConversationSummary } from '@/types/assistant.type';

defineProps<{
  conversations: ConversationSummary[];
  activeId: string;
}>();

const emit = defineEmits<{
  (e: 'select', id: string): void;
  (e: 'delete', id: string): void;
}>();

const formatDate = (isoString?: string) => {
  if (!isoString) return '';
  return new Date(isoString).toLocaleDateString([], { month: 'short', day: 'numeric' });
};
</script>

<template>
  <div class="flex flex-column gap-2 pb-3">
    <div 
      v-for="conv in conversations" 
      :key="conv.id"
      class="history-item flex align-items-center justify-content-between p-2 mx-2 border-round cursor-pointer transition-colors"
      :class="conv.id === activeId ? 'surface-hover text-primary font-medium' : 'hover:surface-hover text-color'"
      @click="emit('select', conv.id)"
    >
      <div class="flex flex-column overflow-hidden mr-2 flex-1 min-w-0">
        <span class="white-space-nowrap overflow-hidden text-overflow-ellipsis text-sm">{{ conv.title || 'New Conversation' }}</span>
        <span class="text-xs text-color-secondary mt-1">{{ formatDate(conv.created_at) }}</span>
      </div>
      <div class="flex-shrink-0">
        <Button 
          icon="pi pi-trash" 
          severity="danger" 
          text 
          rounded 
          size="small"
          class="h-2rem w-2rem delete-btn"
          @click.stop="emit('delete', conv.id)" 
          aria-label="Delete Conversation"
        />
      </div>
    </div>
    
    <div v-if="conversations.length === 0" class="text-center text-color-secondary p-3 text-sm">
      No recent conversations
    </div>
  </div>
</template>

<style scoped>
.history-item .delete-btn {
  opacity: 0;
  transition: opacity 0.2s;
}

.history-item:hover .delete-btn {
  opacity: 1;
}

/* Ensure the delete button is always visible on touch devices if hover media isn't supported */
@media (hover: none) {
  .history-item .delete-btn {
    opacity: 1;
  }
}
</style>
