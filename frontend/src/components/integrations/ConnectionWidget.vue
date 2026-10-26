<script setup lang="ts">
  // Presentation only: typed saved data becomes a summary; actions and persistence belong to the page.
  import { computed } from 'vue';
  import { RouteName } from '@/types/router';
  import type { ConnectionInstance } from '@/types/integrations.type';
  import { connectionWidget } from '@/domain/connection-board';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import WasteBinIcon from './WasteBinIcon.vue';

  const props = defineProps<{
    item: ConnectionInstance;
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
</script>

<template>
  <article
    class="connection-widget"
    :class="{ 'is-editing': editing }"
    :aria-label="item.name"
    :data-instance-id="item.id">
    <header class="connection-widget-head">
      <span class="connection-widget-badge" :data-connection-icon="summary.isWaste ? 'schedule' : 'connections'">
        <UiIcon :name="summary.isWaste ? 'schedule' : 'connections'" />
      </span>
      <div class="connection-widget-title">
        <h2>
          <span v-if="editing">{{ item.name }}</span>
          <RouterLink
            v-else
            :to="{ name: RouteName.Integration, params: { id: item.id } }"
            class="connection-widget-link"
            :data-instance-id="item.id"
            :aria-label="`Open ${item.name} integration`"
            @click="$emit('open', $event)"
            >{{ item.name }}</RouterLink
          >
        </h2>
        <p v-if="item.postcode" class="connection-widget-location">{{ item.postcode }}</p>
      </div>
      <UiButton
        v-if="!editing"
        variant="ghost"
        :icon="item.operation_id ? 'interview' : 'refresh'"
        :label="`${item.operation_id ? 'Check status for' : 'Refresh'} ${item.name}`"
        :disabled="busy || !item.enabled"
        class="connection-widget-action"
        @click="$emit('refresh')" />
    </header>
    <ul v-if="summary.rows.length" class="connection-widget-collections">
      <li v-for="row in summary.rows" :key="row.id">
        <span class="connection-widget-service"
          ><span class="connection-widget-bin"><WasteBinIcon :colour="row.colour" /></span
          ><span>{{ row.name }}</span></span
        >
        <span class="connection-widget-date">{{ row.date }}</span>
      </li>
    </ul>
    <p v-if="summary.remaining" class="connection-widget-meta">+{{ summary.remaining }} more collections</p>
    <div class="connection-widget-meta" :class="{ 'has-error': !!item.last_error }">
      <p>{{ summary.status }}</p>
      <p v-if="summary.checked">Checked {{ summary.checked }}</p>
    </div>
    <div v-if="editing" class="connection-widget-layout">
      <UiButton
        variant="ghost"
        icon="drag"
        :label="`Drag ${item.name}`"
        :disabled="busy"
        class="connection-widget-drag" />
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
    </div>
  </article>
</template>

<style scoped>
  .connection-widget {
    position: relative;
    isolation: isolate;
    min-width: 0;
    padding: 1rem;
    border: var(--nh-border-w) solid var(--nh-border);
    border-radius: var(--nh-radius-l);
    background: var(--nh-surface);
    box-shadow: var(--nh-shadow-1);
  }
  .connection-widget.is-editing {
    border-style: dashed;
  }
  .connection-widget-head {
    display: flex;
    align-items: center;
    gap: 0.65rem;
    margin-bottom: 1rem;
  }
  .connection-widget-badge,
  .connection-widget-bin {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex: none;
    border-radius: 50%;
    background: var(--nh-surface-2);
  }
  .connection-widget-badge {
    width: 3rem;
    height: 3rem;
    color: var(--nh-accent);
  }
  .connection-widget-badge :deep(.nh-ic) {
    width: 1.5rem;
    height: 1.5rem;
  }
  .connection-widget-title {
    min-width: 0;
    flex: 1;
  }
  .connection-widget-title h2 {
    font-size: 0.95rem;
    line-height: 1.4;
    margin: 0;
    overflow-wrap: anywhere;
  }
  .connection-widget-link {
    text-decoration: none;
  }
  /* The real route link covers content; sibling controls stay above its hit area. */
  .connection-widget-link::after {
    content: '';
    position: absolute;
    inset: 0;
    border-radius: var(--nh-radius-l);
    z-index: 1;
  }
  .connection-widget-link:hover::after {
    box-shadow: inset 0 0 0 1px var(--nh-accent);
  }
  .connection-widget-link:focus-visible::after {
    outline: 2px solid var(--nh-focus);
    outline-offset: 3px;
  }
  .connection-widget-location {
    margin: 0.15rem 0 0;
    color: var(--nh-text-muted);
    font-size: 0.8rem;
  }
  .connection-widget-action {
    position: relative;
    z-index: 2;
    flex: none;
    align-self: flex-start;
    margin: -0.35rem -0.4rem 0 0;
  }
  .connection-widget-collections {
    list-style: none;
    display: grid;
    gap: 0.55rem;
    margin: 0;
    padding: 0;
  }
  .connection-widget-collections li {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
    font-size: 0.875rem;
  }
  .connection-widget-service {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    min-width: 0;
    overflow-wrap: anywhere;
  }
  .connection-widget-bin {
    width: 2rem;
    height: 2rem;
  }
  .connection-widget-bin :deep(.waste-bin-icon) {
    width: 1.25rem;
    height: 1.65rem;
  }
  .connection-widget-date {
    text-align: end;
    flex: none;
    max-width: 50%;
    font-weight: 600;
  }
  .connection-widget-meta {
    color: var(--nh-text-muted);
    font-size: 0.8rem;
    margin: 1rem 0 0;
    overflow-wrap: anywhere;
  }
  .connection-widget-meta p {
    margin: 0.15rem 0 0;
  }
  .connection-widget-meta.has-error {
    color: var(--nh-danger);
  }
  .connection-widget-layout {
    position: relative;
    z-index: 2;
    display: flex;
    justify-content: space-between;
    gap: 0.15rem;
    padding-top: 0.75rem;
    margin-top: 1rem;
    border-top: 1px dashed var(--nh-border);
  }
  .connection-widget-drag {
    touch-action: none;
    cursor: grab;
  }
</style>
