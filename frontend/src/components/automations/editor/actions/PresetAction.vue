<script setup lang="ts">
  /* "Cycle presets" action: each run moves the property to its next preset value. */
  import { computed } from 'vue';
  import { Device } from '@/types/device';
  import { AutomationPresetCyclingAction } from '@/types/automation.type';
  import { exposePresets } from '@/domain/exposes';
  import { Issue } from '@/domain/automation';
  import UiSelect from '@/components/ui/UiSelect.vue';
  import IssueText from '../IssueText.vue';
  import { exposeOptions, hasPresets } from '../options';

  const props = defineProps<{ target: Device | undefined; path: string; issues: Issue[]; automationDevice: Device | undefined }>();
  const action = defineModel<AutomationPresetCyclingAction>({ required: true });

  const presets = computed(() => {
    const e = props.target?.exposes[action.value.property];
    return e ? exposePresets(e).map((p) => p.label) : [];
  });
</script>

<template>
  <div class="nh-act-body">
    <div class="nh-set-row">
      <UiSelect v-model="action.property" :options="exposeOptions(target, hasPresets, action.property)" label="Property with presets" placeholder="Choose a property" :invalid="issues.some((i) => i.path === `${path}.property`)" />
    </div>
    <IssueText :issues="issues" :path="`${path}.property`" />
    <p v-if="presets.length" class="nh-help">Each run moves to the next of: {{ presets.join(', ') }}.</p>
  </div>
</template>
