/*
 * Editable copy of one automation (spec 007 US-04, AC-10…AC-15). The draft is a JSON copy of
 * the stored automation, so saving without edits sends exactly what the hub sent (AC-14).
 * New automations start from the triggering device. Validation runs on every change.
 */
import { computed, ref, watch } from 'vue';
import { Automation } from '@/types/automation.type';
import { cloneAutomation, newAutomation, sameAutomation, validateAutomation } from '@/domain/automation';
import { findDevice } from './useHub';
import { useAutomations } from './useAutomations';

export function useAutomationDraft(id: () => string) {
  const { find, save: saveAutomation, remove: removeAutomation, loaded } = useAutomations();

  const stored = computed(() => find(id()));
  const isNew = computed(() => !stored.value);
  const draft = ref<Automation | null>(null);
  /** What the draft is compared with: the stored automation, or the blank new one. */
  const baseline = ref<Automation | null>(null);

  function load() {
    const source = stored.value ?? (findDevice(id()) ? newAutomation(findDevice(id())!) : null);
    baseline.value = source ? cloneAutomation(source) : null;
    draft.value = source ? cloneAutomation(source) : null;
  }

  const dirty = computed(() => !!draft.value && (isNew.value || !baseline.value || !sameAutomation(draft.value, baseline.value)));

  // (re)load when the id changes, when automations arrive, or when the hub updates this one
  // while there are no local edits
  watch([id, stored, loaded, () => !!findDevice(id())], () => {
    if (!draft.value || !dirty.value || draft.value.id !== id()) load();
  }, { immediate: true });

  const validation = computed(() => (draft.value ? validateAutomation(draft.value, findDevice) : { errors: [], warnings: [] }));

  return {
    draft,
    isNew,
    dirty,
    validation,
    discard: load,
    /** Sends the draft to the hub. Returns false when validation errors block saving. */
    save(): boolean {
      if (!draft.value || validation.value.errors.length) return false;
      const payload = cloneAutomation(draft.value);
      saveAutomation(payload);
      baseline.value = cloneAutomation(payload);
      return true;
    },
    remove() {
      if (draft.value && !isNew.value) removeAutomation(draft.value.id);
    },
  };
}
