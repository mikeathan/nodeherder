<script setup lang="ts">
  /*
   * One THEN action. The type picks the editor (strategy map); changing type or device
   * starts that action afresh, as the previous editor did, so no stale fields are sent.
   */
  import { computed } from 'vue';
  import { Device } from '@/types/device';
  import { ActionType, AutomationAction, AutomationActionTypes } from '@/types/automation.type';
  import { createActionFromType } from '@/contracts/automations';
  import { Issue } from '@/domain/automation';
  import { useHub } from '@/composables/useHub';
  import UiSelect from '@/components/ui/UiSelect.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import IssueText from './IssueText.vue';
  import SetValuesAction from './actions/SetValuesAction.vue';
  import StepAction from './actions/StepAction.vue';
  import PresetAction from './actions/PresetAction.vue';
  import { deviceOptions, ExposeFilter, hasPresets, numericWritable, writable } from './options';

  defineProps<{ automationDevice: Device | undefined; path: string; issues: Issue[]; index: number; count: number }>();
  defineEmits<{ (e: 'move', to: number): void; (e: 'remove'): void }>();
  const action = defineModel<AutomationAction>({ required: true });

  const { devices, findDevice } = useHub();

  const STRATEGIES: Record<ActionType, { label: string; component: unknown; targets: ExposeFilter }> = {
    [AutomationActionTypes.Trigger]: { label: 'Set values', component: SetValuesAction, targets: writable },
    [AutomationActionTypes.Step]: { label: 'Change a number by steps', component: StepAction, targets: numericWritable },
    [AutomationActionTypes.PresetCycling]: { label: 'Cycle through presets', component: PresetAction, targets: hasPresets },
  };
  const typeOptions = Object.entries(STRATEGIES).map(([value, s]) => ({ value, label: s.label }));
  const strategy = computed(() => STRATEGIES[action.value.type] ?? STRATEGIES.trigger);
  const target = computed(() => (action.value.id ? findDevice(action.value.id) : undefined));
  const targets = computed(() => deviceOptions(devices.value, strategy.value.targets, action.value.id));

  function setType(type: unknown) {
    if (type === action.value.type) return;
    action.value = { ...createActionFromType(type as ActionType), id: action.value.id };
  }
  function setTarget(id: unknown) {
    if (id === action.value.id) return;
    action.value = { ...createActionFromType(action.value.type), id: String(id) };
  }
</script>

<template>
  <div class="nh-act">
    <div class="nh-act-head">
      <span class="nh-act-n" aria-hidden="true">{{ index + 1 }}</span>
      <UiSelect :model-value="action.type" :options="typeOptions" :label="`Action ${index + 1} type`" @update:model-value="setType" />
      <span class="nh-reorder">
        <UiButton size="sm" variant="ghost" icon="moveUp" label="Move action up" :disabled="index === 0" @click="$emit('move', index - 1)" />
        <UiButton size="sm" variant="ghost" icon="moveDown" label="Move action down" :disabled="index === count - 1" @click="$emit('move', index + 1)" />
        <UiButton size="sm" variant="ghost" icon="delete" label="Remove action" @click="$emit('remove')" />
      </span>
    </div>
    <div class="nh-set-row">
      <UiSelect :model-value="action.id" :options="targets" :label="`Action ${index + 1} device`" placeholder="Choose a device to control" :invalid="issues.some((i) => i.path === `${path}.id`)" @update:model-value="setTarget" />
    </div>
    <IssueText :issues="issues" :path="`${path}.id`" />
    <component :is="strategy.component" v-if="action.id" v-model="action" :target="target" :automation-device="automationDevice" :path="path" :issues="issues" />
  </div>
</template>
