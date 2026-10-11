<script setup lang="ts">
  /* One IF condition: a device value comparison or a time window. */
  import { computed } from 'vue';
  import { Device } from '@/types/device';
  import { AutomationCondition, isExposeCondition, isTimeCondition } from '@/types/automation.type';
  import { Issue } from '@/domain/automation';
  import UiSelect from '@/components/ui/UiSelect.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import ExposeValueInput from './ExposeValueInput.vue';
  import IssueText from './IssueText.vue';
  import { comparisonOptions, exposeOptions, notConfig } from './options';

  const props = defineProps<{ device: Device | undefined; path: string; issues: Issue[]; index: number; count: number }>();
  defineEmits<{ (e: 'move', to: number): void; (e: 'remove'): void }>();
  const condition = defineModel<AutomationCondition>({ required: true });

  const has = (path: string) => props.issues.some((i) => i.path === `${props.path}.${path}`);
  const expose = computed(() => (isExposeCondition(condition.value) ? props.device?.exposes[condition.value.name] : undefined));
  const ops = computed(() => comparisonOptions(expose.value));

  function setProperty(name: unknown) {
    const c = condition.value;
    if (!isExposeCondition(c) || c.name === name) return;
    c.name = String(name);
    c.value = null;
    if (!comparisonOptions(props.device?.exposes[c.name]).some((o) => o.value === c.equality)) c.equality = '=';
  }
</script>

<template>
  <div class="nh-cond">
    <div class="nh-cond-head">
      <span class="nh-cond-kind"><UiIcon :name="isTimeCondition(condition) ? 'clock' : 'gauge'" />{{ isTimeCondition(condition) ? 'Time' : 'Value' }}</span>
      <span class="nh-reorder">
        <UiButton size="sm" variant="ghost" icon="moveUp" label="Move condition up" :disabled="index === 0" @click="$emit('move', index - 1)" />
        <UiButton size="sm" variant="ghost" icon="moveDown" label="Move condition down" :disabled="index === count - 1" @click="$emit('move', index + 1)" />
        <UiButton size="sm" variant="ghost" icon="delete" label="Remove condition" @click="$emit('remove')" />
      </span>
    </div>
    <div v-if="isExposeCondition(condition)" class="nh-cond-body">
      <UiSelect :model-value="condition.name" :options="exposeOptions(device, notConfig, condition.name)" label="Property" placeholder="Choose a property" :invalid="has('name')" @update:model-value="setProperty" />
      <UiSelect v-model="condition.equality" class="is-op" :options="ops" label="Comparison" :disabled="!condition.name" />
      <ExposeValueInput v-model="condition.value" :expose="expose" label="Compare with" :invalid="has('value')" :disabled="!condition.name" />
    </div>
    <div v-else-if="isTimeCondition(condition)" class="nh-cond-body">
      <label class="nh-time">From <input v-model="condition.timeRange.startAt" type="time" class="nh-input" :aria-invalid="has('timeRange.startAt') || undefined" /></label>
      <label class="nh-time">to <input v-model="condition.timeRange.endAt" type="time" class="nh-input" :aria-invalid="has('timeRange.endAt') || undefined" /></label>
      <small v-if="condition.timeRange.startAt && condition.timeRange.endAt && condition.timeRange.startAt > condition.timeRange.endAt" class="nh-muted">Runs overnight</small>
    </div>
    <IssueText :issues="issues" :path="`${path}.name`" />
    <IssueText :issues="issues" :path="`${path}.value`" />
    <IssueText :issues="issues" :path="`${path}.timeRange.startAt`" />
    <IssueText :issues="issues" :path="`${path}.timeRange.endAt`" />
  </div>
</template>

<style scoped>
  .nh-cond {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    padding: 0.55rem 0.65rem;
    border-radius: var(--nh-radius-m);
    background: var(--nh-surface-2);
    border: var(--nh-border-w) solid var(--nh-border);
    min-width: 0;
  }
  .nh-cond-head {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }
  .nh-cond-kind {
    display: inline-flex;
    align-items: center;
    gap: 0.3rem;
    font-size: 0.72rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--nh-text-muted);
  }
  .nh-reorder {
    display: inline-flex;
    margin-left: auto;
  }
  .nh-cond-body {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.45rem;
    min-width: 0;
  }
  .nh-cond-body > .nh-select:first-child {
    flex: 1 1 11rem;
  }
  .nh-cond-body > .is-op {
    width: 7.5rem;
  }
  .nh-time {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    font-size: 0.875rem;
  }
  @media (max-width: 640px) {
    .nh-cond-body > .nh-select:first-child {
      flex-basis: 100%;
    }
    .nh-cond-body > .is-op {
      flex: 0 0 7rem;
    }
    .nh-cond-body > :last-child:not(.nh-select:first-child) {
      flex: 1 1 8rem;
    }
    .nh-time {
      flex: 1 1 8rem;
    }
    .nh-time input {
      flex: 1;
      min-width: 0;
    }
  }
</style>
