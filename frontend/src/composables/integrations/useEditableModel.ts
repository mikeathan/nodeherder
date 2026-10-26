import { ref, toRaw, watch, type Ref } from 'vue';

/** Clone JSON definition data; it contains no Dates, functions or runtime secrets. */
export function cloneDefinitionValue<T>(value: T): T {
  return JSON.parse(JSON.stringify(value)) as T;
}

/**
 * Gives a form section an owned draft instead of mutating nested parent props.
 * Parent replacements reset the section; the section's own emitted value is not echoed back.
 */
export function useEditableModel<T>(read: () => T, emit: (value: T) => void): Ref<T> {
  const draft = ref(cloneDefinitionValue(read())) as Ref<T>;
  let emitted: T | undefined;
  let replacing = false;
  watch(
    read,
    (value) => {
      if (toRaw(value) === toRaw(emitted)) return;
      replacing = true;
      draft.value = cloneDefinitionValue(value);
      replacing = false;
    },
    { flush: 'sync' }
  );
  watch(
    draft,
    (value) => {
      if (replacing) return;
      emitted = cloneDefinitionValue(value);
      emit(emitted);
    },
    { deep: true, flush: 'sync' }
  );
  return draft;
}
