<script setup lang="ts">
  /*
   * Automation editor (spec 007 US-04, FR-06): the whole automation on one page instead of
   * the previous stack of panels. Edits stay local until Save; leaving with unsaved changes
   * asks first. Saving sends the existing saveAutomation command with the edited payload.
   */
  import { computed, ref } from 'vue';
  import { onBeforeRouteLeave, useRouter } from 'vue-router';
  import { RouteName } from '@/types/router';
  import { canTriggerManually } from '@/contracts/automations';
  import { moveItem } from '@/domain/dashboard';
  import { newTrigger } from '@/domain/automation';
  import { useAutomationDraft } from '@/composables/useAutomationDraft';
  import { useAutomations } from '@/composables/useAutomations';
  import { findDevice, useHub } from '@/composables/useHub';
  import { confirm } from '@/composables/useConfirm';
  import UiPageHeader from '@/components/ui/UiPageHeader.vue';
  import UiCard from '@/components/ui/UiCard.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiEmpty from '@/components/ui/UiEmpty.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import UiField from '@/components/ui/UiField.vue';
  import TriggerCard from './TriggerCard.vue';
  import ScheduleCard from './ScheduleCard.vue';

  const props = defineProps<{ id: string }>();
  const router = useRouter();
  const { initialized } = useHub();
  const { loaded, run } = useAutomations();
  const { draft, isNew, dirty, validation, discard, save, remove } = useAutomationDraft(() => props.id);

  const device = computed(() => findDevice(props.id));
  const errors = computed(() => validation.value.errors);
  const warnings = computed(() => validation.value.warnings);
  const triedToSave = ref(false);
  let leaving = false;

  function onSave() {
    triedToSave.value = true;
    if (!save()) return;
    leaving = true;
    router.push({ name: RouteName.Viewer });
  }

  async function onDelete() {
    if (!draft.value) return;
    const ok = await confirm({ title: `Delete “${draft.value.friendlyname}”?`, message: 'The automation and all its triggers are removed from the hub.', confirmLabel: 'Delete', danger: true });
    if (!ok) return;
    remove();
    leaving = true;
    router.push({ name: RouteName.Viewer });
  }

  async function removeTrigger(index: number) {
    const ok = await confirm({ title: `Delete trigger ${index + 1}?`, message: 'Its conditions and actions are removed when you save.', confirmLabel: 'Delete trigger', danger: true });
    if (ok && draft.value) draft.value.triggers.splice(index, 1);
  }

  function addTrigger() {
    draft.value?.triggers.push(newTrigger());
  }

  function moveTrigger(from: number, to: number) {
    if (draft.value) draft.value.triggers = moveItem(draft.value.triggers, from, to);
  }

  onBeforeRouteLeave(async () => {
    if (leaving || !dirty.value) return true;
    return confirm({ title: 'Leave without saving?', message: 'Your changes to this automation will be lost.', confirmLabel: 'Leave', danger: true });
  });

  const json = computed(() => JSON.stringify(draft.value, null, 2));
</script>

<template>
  <div v-if="draft" class="nh-editor">
    <div class="nh-editor-main">
      <UiPageHeader :title="draft.friendlyname || draft.id" :back="{ name: RouteName.Viewer }" back-label="Automations">
        <template #subtitle>
          {{ isNew ? 'New automation for' : 'Started by' }} <RouterLink :to="{ name: RouteName.DevicePage, params: { id: draft.id } }">{{ device?.friendly_name ?? draft.id }}</RouterLink>
        </template>
      </UiPageHeader>

      <UiCard title="Details" icon="info">
        <UiField label="Description" for="auto-desc" help="Optional. Shown in the automation list.">
          <textarea id="auto-desc" v-model="draft.description" class="nh-input" rows="2" />
        </UiField>
      </UiCard>

      <TriggerCard
        v-for="(t, i) in draft.triggers"
        :key="i"
        v-model="draft.triggers[i]"
        :automation="draft"
        :automation-device="device"
        :path="`triggers.${i}`"
        :errors="errors"
        :warnings="warnings"
        :index="i"
        :count="draft.triggers.length"
        :can-run="!isNew && !dirty && canTriggerManually(t)"
        @move="moveTrigger(i, $event)"
        @remove="removeTrigger(i)"
        @run="run(draft, t.name)" />

      <button type="button" class="nh-add-trig" @click="addTrigger"><UiIcon name="add" />Add trigger</button>
    </div>

    <aside class="nh-editor-side">
      <UiCard title="Check" icon="check">
        <p v-if="errors.length" class="nh-sum is-err"><UiIcon name="error" />{{ errors.length }} thing{{ errors.length === 1 ? '' : 's' }} to fix before saving</p>
        <p v-else class="nh-sum is-ok"><UiIcon name="check" />Ready to save</p>
        <p v-if="warnings.length" class="nh-sum is-warn"><UiIcon name="warn" />{{ warnings.length }} warning{{ warnings.length === 1 ? '' : 's' }}</p>
        <ul v-if="triedToSave && errors.length" class="nh-sum-list">
          <li v-for="e in errors" :key="e.path">{{ e.message }}</li>
        </ul>
      </UiCard>
      <ScheduleCard v-model="draft" />
      <details class="nh-card nh-json">
        <summary><UiIcon name="json" />Payload sent to the hub</summary>
        <pre>{{ json }}</pre>
      </details>
    </aside>

    <div class="nh-savebar" role="region" aria-label="Save changes">
      <span class="nh-dirty">
        <template v-if="isNew">New automation · not saved yet</template>
        <template v-else-if="dirty"><span class="nh-dot" aria-hidden="true" />Unsaved changes</template>
        <template v-else>No changes</template>
      </span>
      <UiButton v-if="!isNew" variant="ghost" icon="delete" class="is-danger" @click="onDelete">Delete</UiButton>
      <UiButton variant="ghost" icon="refresh" :disabled="!dirty || isNew" @click="discard">Discard</UiButton>
      <UiButton variant="primary" icon="check" :disabled="!!errors.length && triedToSave" @click="onSave">Save</UiButton>
    </div>
  </div>
  <div v-else-if="!initialized || !loaded" class="nh-card nh-editor-skel" aria-busy="true"><span class="nh-skel" /><span class="nh-skel" /></div>
  <UiEmpty v-else icon="automation" title="Automation not found" text="There is no automation or device with this id.">
    <UiButton :to="{ name: RouteName.Viewer }" icon="back">Back to automations</UiButton>
  </UiEmpty>
</template>

<style scoped>
  .nh-editor {
    display: grid;
    grid-template-columns: minmax(0, 1fr) 19rem;
    gap: calc(var(--nh-space) * 1.2);
    align-items: start;
  }
  .nh-editor-main {
    display: flex;
    flex-direction: column;
    gap: var(--nh-space);
    min-width: 0;
  }
  .nh-editor-main :deep(.nh-card + .nh-card) {
    margin-top: 0;
  }
  .nh-editor-side {
    position: sticky;
    top: calc(var(--nh-header-h) + 0.5rem);
    display: flex;
    flex-direction: column;
    gap: var(--nh-space);
    min-width: 0;
  }
  .nh-editor-side :deep(.nh-card + .nh-card) {
    margin-top: 0;
  }
  .nh-sum {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    margin: 0 0 0.3rem;
    font-size: 0.875rem;
  }
  .nh-sum.is-err .nh-ic {
    color: var(--nh-danger);
  }
  .nh-sum.is-ok .nh-ic {
    color: var(--nh-ok);
  }
  .nh-sum.is-warn .nh-ic {
    color: var(--nh-warn);
  }
  .nh-sum-list {
    margin: 0.4rem 0 0;
    padding-left: 1.1rem;
    font-size: 0.82rem;
    color: var(--nh-danger);
  }
  .nh-json summary {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem var(--nh-space);
    cursor: pointer;
    font-weight: 600;
    font-size: 0.9rem;
  }
  .nh-json pre {
    margin: 0;
    padding: var(--nh-space);
    max-height: 26rem;
    overflow: auto;
    font: 0.75rem/1.5 var(--nh-font-mono);
    background: var(--nh-surface-2);
  }
  .nh-add-trig {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    width: 100%;
    padding: 1rem;
    border: 2px dashed var(--nh-border-strong);
    border-radius: var(--nh-radius-l);
    background: transparent;
    color: var(--nh-text-muted);
    font-weight: 600;
    cursor: pointer;
  }
  .nh-add-trig:hover {
    color: var(--nh-accent);
    border-color: var(--nh-accent);
  }
  .nh-savebar {
    position: sticky;
    bottom: 0.75rem;
    grid-column: 1 / -1;
    z-index: 5;
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: flex-end;
    gap: 0.5rem;
    padding: 0.65rem var(--nh-space);
    background: var(--nh-surface);
    border: var(--nh-border-w) solid var(--nh-border-strong);
    border-radius: var(--nh-radius-l);
    box-shadow: var(--nh-shadow-2);
  }
  .nh-dirty {
    margin-right: auto;
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    color: var(--nh-text-muted);
    font-size: 0.85rem;
  }
  .nh-dot {
    width: 0.55rem;
    height: 0.55rem;
    border-radius: 50%;
    background: var(--nh-warn);
  }
  .nh-editor-skel {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    padding: 1rem;
  }
  @media (max-width: 1100px) {
    .nh-editor {
      grid-template-columns: minmax(0, 1fr);
    }
    .nh-editor-side {
      position: static;
    }
  }
  @media (max-width: 640px) {
    .nh-savebar {
      padding: 0.5rem 0.6rem;
      border-radius: var(--nh-radius-m);
      bottom: 0.5rem;
    }
    .nh-savebar .nh-dirty {
      flex: 1 1 100%;
    }
    .nh-savebar > :deep(.nh-btn) {
      flex: 1 1 auto;
    }
  }
</style>

<style>
  /* Shared by the editor's child components (trigger, condition and action editors). */
  .nh-trig {
    background: var(--nh-surface);
    border: var(--nh-border-w) solid transparent;
    border-radius: var(--nh-radius-l);
    box-shadow: var(--nh-shadow-1);
    min-width: 0;
  }
  [data-mode='dark'] .nh-trig {
    border-color: var(--nh-border);
  }
  .nh-trig.has-err {
    border-left: 3px solid var(--nh-danger);
  }
  .nh-trig-head {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.4rem 0.75rem;
    padding: 0.6rem var(--nh-space);
    border-bottom: var(--nh-border-w) solid var(--nh-border);
  }
  .nh-trig.is-collapsed .nh-trig-head {
    border-bottom: 0;
  }
  .nh-trig-toggle {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    border: 0;
    background: none;
    color: inherit;
    cursor: pointer;
    padding: 0.2rem;
    min-height: 2rem;
  }
  .nh-trig-tools {
    display: flex;
    align-items: center;
    gap: 0.2rem;
    margin-left: auto;
  }
  .nh-sentence {
    flex-basis: 100%;
    margin: 0;
    font-size: 0.875rem;
    color: var(--nh-text-muted);
    font-style: italic;
  }
  .nh-flow {
    padding: var(--nh-space);
  }
  .nh-block {
    display: grid;
    grid-template-columns: 4.5rem minmax(0, 1fr);
    position: relative;
  }
  .nh-block-label {
    position: relative;
    display: flex;
    justify-content: center;
  }
  .nh-block-label::after {
    content: '';
    position: absolute;
    top: 2.2rem;
    bottom: 0;
    width: 2px;
    background: var(--nh-border-strong);
  }
  .nh-block:last-child .nh-block-label::after {
    display: none;
  }
  .nh-block-label span {
    position: relative;
    z-index: 1;
    display: grid;
    place-items: center;
    height: 2rem;
    padding: 0 0.6rem;
    border-radius: 99px;
    font-size: 0.72rem;
    font-weight: 800;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    background: var(--nh-bc);
    color: var(--nh-bc-text, #fff);
  }
  .nh-block.is-when {
    --nh-bc: var(--nh-accent);
    --nh-bc-text: var(--nh-accent-contrast);
  }
  .nh-block.is-if {
    --nh-bc: var(--nh-info);
  }
  .nh-block.is-then {
    --nh-bc: var(--nh-ok);
  }
  .nh-block-body {
    display: flex;
    flex-direction: column;
    gap: 0.55rem;
    padding: 0.15rem 0 1.3rem;
    min-width: 0;
  }
  .nh-block:last-child .nh-block-body {
    padding-bottom: 0.2rem;
  }
  .nh-block-note {
    margin: 0;
    font-size: 0.85rem;
  }
  .nh-src {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    font-weight: 600;
    padding: 0 0.6rem;
    min-height: var(--nh-ctl-h);
    border-radius: var(--nh-radius-s);
    background: var(--nh-surface-2);
  }
  .nh-act {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    padding: 0.6rem 0.65rem;
    border-radius: var(--nh-radius-m);
    background: var(--nh-surface-2);
    border: var(--nh-border-w) solid var(--nh-border);
    min-width: 0;
  }
  .nh-act-head {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.5rem;
  }
  .nh-act-head > .nh-select {
    flex: 0 1 15rem;
  }
  .nh-act-n {
    display: grid;
    place-items: center;
    width: 1.6rem;
    height: 1.6rem;
    border-radius: 50%;
    background: var(--nh-ok);
    color: #fff;
    font-size: 0.75rem;
    font-weight: 700;
    flex: none;
  }
  .nh-act-body {
    display: flex;
    flex-direction: column;
    gap: 0.45rem;
    padding-left: 2.1rem;
    min-width: 0;
  }
  .nh-set-row {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.4rem;
    min-width: 0;
  }
  .nh-set-row > .nh-select:first-child {
    flex: 1 1 12rem;
  }
  .nh-act > .nh-set-row > .nh-select {
    flex: 1 1 14rem;
    max-width: 24rem;
  }
  .nh-act-opts {
    padding-top: 0.2rem;
  }
  .nh-arrow {
    color: var(--nh-text-muted);
  }
  .nh-reorder {
    display: inline-flex;
    margin-left: auto;
  }
  .nh-warn-list {
    list-style: none;
    margin: 0;
    padding: 0.6rem var(--nh-space);
    border-top: var(--nh-border-w) solid var(--nh-border);
    background: var(--nh-warn-soft);
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
    font-size: 0.82rem;
    border-radius: 0 0 var(--nh-radius-l) var(--nh-radius-l);
  }
  .nh-warn-list li {
    display: flex;
    gap: 0.4rem;
    align-items: flex-start;
  }
  .nh-warn-list .nh-ic {
    color: var(--nh-warn);
    margin-top: 0.1rem;
  }
  @media (max-width: 700px) {
    .nh-flow {
      padding: 0.75rem;
    }
    .nh-block {
      grid-template-columns: minmax(0, 1fr);
    }
    .nh-block-label {
      justify-content: flex-start;
      margin-bottom: 0.45rem;
    }
    .nh-block-label::after {
      display: none;
    }
    .nh-block-label span {
      height: 1.6rem;
      font-size: 0.68rem;
    }
    .nh-block-body {
      margin-left: 0.7rem;
      padding: 0 0 1rem 0.8rem;
      border-left: 2px solid var(--nh-border-strong);
    }
    .nh-act-body {
      padding-left: 0;
    }
    .nh-act-head > .nh-select,
    .nh-act > .nh-set-row > .nh-select,
    .nh-set-row > .nh-select:first-child,
    .nh-src {
      flex: 1 1 100%;
      max-width: none;
    }
    .nh-act-head > .nh-select {
      order: 3;
    }
    .nh-trig-head {
      padding: 0.6rem 0.75rem;
    }
    .nh-trig-tools {
      flex-wrap: wrap;
      justify-content: flex-end;
    }
  }
</style>
