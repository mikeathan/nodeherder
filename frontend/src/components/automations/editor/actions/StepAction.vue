<script setup lang="ts">
  /*
   * "Change by steps" action: target = step₀ op₀ (step₁ op₁ (… amount)), computed by the hub
   * from current values (backend operations.go). The first step is always the target's own
   * value; more steps can read the target or the triggering device.
   */
  import { computed } from 'vue';
  import { Device } from '@/types/device';
  import { AutomationStepAction, NumericOperator } from '@/types/automation.type';
  import { Issue, stepFormula } from '@/domain/automation';
  import UiSelect from '@/components/ui/UiSelect.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import IssueText from '../IssueText.vue';
  import { exposeOptions, numericMeasurement, numericWritable, STEP_OPERATORS } from '../options';

  const props = defineProps<{ target: Device | undefined; path: string; issues: Issue[]; automationDevice: Device | undefined }>();
  const action = defineModel<AutomationStepAction>({ required: true });

  const has = (path: string) => props.issues.some((i) => i.path === `${props.path}.${path}`);
  const sources = computed(() =>
    [props.target, props.automationDevice]
      .filter((d): d is Device => !!d)
      .filter((d, i, all) => all.findIndex((x) => x.id === d.id) === i)
      .map((d) => ({ value: d.id, label: d.friendly_name }))
  );
  const sourceDevice = (id: string) => [props.target, props.automationDevice].find((d) => d?.id === id);

  function setTargetProperty(name: unknown) {
    const a = action.value;
    a.property = String(name);
    if (a.steps.length) a.steps[0] = { ...a.steps[0], id: a.id, property: a.property };
    else a.steps.push({ id: a.id, property: a.property, operator: '+' });
  }
  const addStep = () => action.value.steps.push({ id: props.automationDevice?.id ?? action.value.id, property: '', operator: '*' as NumericOperator });
  const removeStep = (i: number) => action.value.steps.splice(i, 1);
  function setAmount(event: Event) {
    const raw = (event.target as HTMLInputElement).value;
    action.value.data = raw === '' ? null : Number(raw);
  }
</script>

<template>
  <div class="nh-act-body">
    <div class="nh-set-row">
      <UiSelect :model-value="action.property" :options="exposeOptions(target, numericWritable, action.property)" label="Number to change" placeholder="Choose a number" :invalid="has('property')" @update:model-value="setTargetProperty" />
    </div>
    <IssueText :issues="issues" :path="`${path}.property`" />
    <ol v-if="action.steps.length" class="nh-steps">
      <li v-for="(step, i) in action.steps" :key="i" class="nh-set-row">
        <UiSelect v-model="step.operator" class="is-op" :options="STEP_OPERATORS" :label="`Step ${i + 1} operation`" />
        <template v-if="i === 0">
          <span class="nh-muted">current {{ target?.friendly_name ?? 'target' }} value</span>
        </template>
        <template v-else>
          <UiSelect :model-value="step.id" :options="sources" :label="`Step ${i + 1} device`" @update:model-value="step.id = String($event); step.property = ''" />
          <UiSelect v-model="step.property" :options="exposeOptions(sourceDevice(step.id), numericMeasurement, step.property)" :label="`Step ${i + 1} value`" placeholder="Choose a value" :invalid="has(`steps.${i}`)" />
          <UiButton size="sm" variant="ghost" icon="remove" :label="`Remove step ${i + 1}`" @click="removeStep(i)" />
        </template>
        <IssueText :issues="issues" :path="`${path}.steps.${i}`" />
      </li>
    </ol>
    <IssueText :issues="issues" :path="`${path}.steps`" />
    <div class="nh-inline">
      <UiButton size="sm" icon="add" :disabled="!action.steps.length" @click="addStep">Add step</UiButton>
      <label class="nh-inline">
        <span class="nh-label">Amount</span>
        <input type="number" step="any" class="nh-input is-num" :value="typeof action.data === 'number' ? action.data : ''" aria-label="Amount" :aria-invalid="has('data') || undefined" @input="setAmount" />
      </label>
    </div>
    <IssueText :issues="issues" :path="`${path}.data`" />
    <p v-if="action.property" class="nh-help">New {{ action.property }} = {{ stepFormula(action) }}</p>
  </div>
</template>

<style scoped>
  .nh-steps {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.45rem;
  }
  .is-op {
    width: 9rem;
  }
</style>
