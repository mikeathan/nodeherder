<script setup lang="ts">
  import { computed } from 'vue';
  import { RouteName } from '@/types/router';
  import type { ConnectionInstance } from '@/types/integrations';
  import { connectionWidget } from '@/domain/integrations/connection-board';
  import { integrationTypeIcon } from '@/domain/integrations/catalog';
  import { formatScheduleInterval } from '@/domain/schedule';
  import { formatLocalDateTime } from '@/utils/date.utils';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import IntegrationStatus from '../shared/IntegrationStatus.vue';

  const props = defineProps<{
    item: ConnectionInstance;
    typeTitle: string;
    providerName?: string;
    editing: boolean;
    busy: boolean;
    first: boolean;
    last: boolean;
  }>();
  defineEmits<{
    open: [event: MouseEvent];
    refresh: [];
    hide: [];
    move: [offset: number];
  }>();
  const summary = computed(() => connectionWidget(props.item));
  const typeId = computed(() => props.item.provider_key.split('/')[0]);
  const schedule = computed(() => {
    const value = props.item.schedule;
    if (!value?.enabled) return 'Manual refresh';
    const interval = `Auto ${formatScheduleInterval(value.interval_seconds)}`;
    return value.next_due_at ? `${interval} · next ${formatLocalDateTime(value.next_due_at)}` : interval;
  });
  const firstRow = computed(() => summary.value.rows[0]);
</script>

<template>
  <article class="integration-entry" :aria-label="item.name" :data-instance-id="item.id">
    <span class="integration-entry-icon">
      <component v-if="summary.icon" :is="summary.icon.component" v-bind="summary.icon.props" />
      <UiIcon v-else :name="integrationTypeIcon(typeId)" />
    </span>
    <div class="integration-entry-main">
      <div class="integration-entry-heading">
        <h3>
          <span v-if="editing">{{ item.name }}</span>
          <RouterLink
            v-else
            :to="{ name: RouteName.Integration, params: { id: item.id } }"
            class="integration-entry-link"
            :data-instance-id="item.id"
            :aria-label="`Open ${item.name} integration`"
            @click="$emit('open', $event)"
            >{{ item.name }}</RouterLink
          >
        </h3>
        <span class="integration-entry-type">{{ providerName || typeTitle }}</span>
      </div>
      <p v-if="firstRow" class="integration-entry-summary">
        <span>{{ firstRow.name }}</span>
        <strong>{{ firstRow.date }}</strong>
        <span v-if="summary.remaining" class="integration-entry-more"
          >+{{ summary.remaining }} {{ summary.remainingLabel }}</span
        >
      </p>
      <p class="integration-entry-schedule">{{ schedule }}</p>
      <IntegrationStatus
        class="integration-entry-status"
        :status="summary.status"
        :checked="summary.checked"
        :error="!!item.last_error"
        :warning="!item.enabled" />
    </div>
    <div class="integration-entry-actions">
      <UiButton
        v-if="!editing"
        variant="ghost"
        :icon="item.operation_id ? 'interview' : 'refresh'"
        :label="`${item.operation_id ? 'Check status for' : 'Refresh'} ${item.name}`"
        :disabled="busy || !item.enabled"
        @click="$emit('refresh')" />
      <template v-if="editing">
        <UiButton
          variant="ghost"
          icon="moveUp"
          :label="`Move ${item.name} earlier`"
          :disabled="busy || first"
          data-move="-1"
          @click="$emit('move', -1)" />
        <UiButton
          variant="ghost"
          icon="moveDown"
          :label="`Move ${item.name} later`"
          :disabled="busy || last"
          data-move="1"
          @click="$emit('move', 1)" />
        <UiButton
          variant="ghost"
          icon="remove"
          :label="`Remove ${item.name} from page`"
          :disabled="busy"
          @click="$emit('hide')" />
      </template>
    </div>
  </article>
</template>

<style scoped>
  .integration-entry {
    display: flex;
    align-items: flex-start;
    gap: 0.8rem;
    min-width: 0;
    padding: 0.9rem 1rem;
    border: var(--nh-border-w) solid var(--nh-border);
    border-radius: var(--nh-radius-l);
    background: var(--nh-surface);
  }
  .integration-entry-icon {
    display: grid;
    place-items: center;
    flex: none;
    width: 2.5rem;
    height: 2.5rem;
    border-radius: 50%;
    background: var(--nh-surface-2);
    color: var(--nh-accent);
  }
  .integration-entry-icon :deep(.nh-ic),
  .integration-entry-icon :deep(svg) {
    width: 1.3rem;
    height: 1.3rem;
  }
  .integration-entry-main {
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
    min-width: 0;
    flex: 1;
  }
  .integration-entry-heading {
    display: flex;
    align-items: baseline;
    flex-wrap: wrap;
    gap: 0.5rem;
  }
  .integration-entry-heading h3 {
    margin: 0;
    font-size: 0.95rem;
    overflow-wrap: anywhere;
  }
  .integration-entry-link {
    text-decoration: none;
  }
  .integration-entry-link:hover {
    text-decoration: underline;
  }
  .integration-entry-type {
    color: var(--nh-text-muted);
    font-size: 0.72rem;
  }
  .integration-entry-summary {
    display: flex;
    flex-wrap: wrap;
    gap: 0.35rem 0.6rem;
    margin: 0.15rem 0 0;
    font-size: 0.84rem;
    overflow-wrap: anywhere;
  }
  .integration-entry-summary strong {
    font-weight: 600;
  }
  .integration-entry-more,
  .integration-entry-schedule {
    color: var(--nh-text-muted);
    font-size: 0.75rem;
  }
  .integration-entry-schedule {
    margin: 0.15rem 0 0;
  }
  .integration-entry-status {
    font-size: 0.75rem;
    margin-top: 0.2rem;
  }
  .integration-entry-actions {
    display: flex;
    align-items: flex-start;
    gap: 0.15rem;
    flex: none;
  }
  @media (max-width: 640px) {
    .integration-entry {
      flex-wrap: wrap;
    }
    .integration-entry-actions {
      width: 100%;
      justify-content: flex-end;
    }
  }
</style>
