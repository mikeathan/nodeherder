<script setup lang="ts">
  /* "Set values" action: one or more expose = value pairs, optional wait, publish mode. */
  import { Device } from '@/types/device';
  import { AutomationTriggerAction, PublishModes } from '@/types/automation.type';
  import { TimeUnit } from '@/types/types.type';
  import { Issue } from '@/domain/automation';
  import UiSelect from '@/components/ui/UiSelect.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiSegmented from '@/components/ui/UiSegmented.vue';
  import ExposeValueInput from '../ExposeValueInput.vue';
  import IssueText from '../IssueText.vue';
  import { exposeOptions, writable } from '../options';

  const props = defineProps<{ target: Device | undefined; path: string; issues: Issue[]; automationDevice: Device | undefined }>();
  const action = defineModel<AutomationTriggerAction>({ required: true });

  const has = (path: string) => props.issues.some((i) => i.path === `${props.path}.${path}`);
  const units: { value: TimeUnit; label: string }[] = [
    { value: 'seconds', label: 'seconds' },
    { value: 'minutes', label: 'minutes' },
    { value: 'hours', label: 'hours' },
  ];
  const modes = [
    { value: PublishModes.Batch, label: 'All at once' },
    { value: PublishModes.Single, label: 'One by one' },
  ];

  function setProperty(row: { name: string; data: unknown }, name: unknown) {
    if (row.name === name) return;
    row.name = String(name);
    row.data = null;
  }
  const addRow = () => action.value.exposes.push({ name: '', data: null });
  const removeRow = (i: number) => action.value.exposes.splice(i, 1);
  function setDelay(event: Event) {
    const raw = (event.target as HTMLInputElement).value;
    action.value.delay = { unit: action.value.delay?.unit ?? 'seconds', value: raw === '' ? 0 : Number(raw) };
  }
</script>

<template>
  <div class="nh-act-body">
    <div v-for="(row, i) in action.exposes" :key="i" class="nh-set-row">
      <UiSelect :model-value="row.name" :options="exposeOptions(target, writable, row.name)" label="Property to set" placeholder="Choose a property" :invalid="has(`exposes.${i}.name`)" @update:model-value="setProperty(row, $event)" />
      <span class="nh-arrow" aria-hidden="true">→</span>
      <ExposeValueInput v-model="row.data" :expose="target?.exposes[row.name]" label="New value" :invalid="has(`exposes.${i}.data`)" :disabled="!row.name" />
      <UiButton size="sm" variant="ghost" icon="remove" label="Remove this value" @click="removeRow(i)" />
      <IssueText :issues="issues" :path="`${path}.exposes.${i}.name`" />
      <IssueText :issues="issues" :path="`${path}.exposes.${i}.data`" />
    </div>
    <IssueText :issues="issues" :path="`${path}.exposes`" />
    <div class="nh-inline">
      <UiButton size="sm" icon="add" :disabled="!action.id" @click="addRow">Add value</UiButton>
    </div>
    <div class="nh-inline nh-act-opts">
      <label class="nh-inline">
        <span class="nh-label">Wait</span>
        <input type="number" min="0" step="1" class="nh-input is-num" :value="action.delay?.value ?? 0" aria-label="Wait before acting" @input="setDelay" />
      </label>
      <UiSelect
        :model-value="action.delay?.unit ?? 'seconds'"
        :options="units"
        label="Wait unit"
        @update:model-value="action.delay = { value: action.delay?.value ?? 0, unit: $event as TimeUnit }" />
      <template v-if="action.exposes.length > 1">
        <span class="nh-label">Send</span>
        <UiSegmented :model-value="action.publishMode ?? PublishModes.Batch" :options="modes" label="How to send several values" @update:model-value="action.publishMode = $event" />
      </template>
    </div>
    <IssueText :issues="issues" :path="`${path}.delay.value`" />
  </div>
</template>
