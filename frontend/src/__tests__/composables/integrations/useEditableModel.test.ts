import { effectScope, shallowRef } from 'vue';
import { useEditableModel } from '@/composables/integrations/useEditableModel';

test('nested form edits emit owned snapshots without mutating parent props or echo loops', () => {
  const source = { auth: { mode: 'none' } };
  const parent = shallowRef(source);
  const emit = jest.fn((value) => {
    parent.value = value;
  });
  const scope = effectScope();
  const draft = scope.run(() => useEditableModel(() => parent.value, emit))!;
  draft.value.auth.mode = 'bearer';
  expect(source.auth.mode).toBe('none');
  expect(parent.value.auth.mode).toBe('bearer');
  expect(emit).toHaveBeenCalledTimes(1);
  parent.value = { auth: { mode: 'optional_bearer' } };
  expect(draft.value.auth.mode).toBe('optional_bearer');
  expect(emit).toHaveBeenCalledTimes(1);
  scope.stop();
  draft.value.auth.mode = 'none';
  expect(emit).toHaveBeenCalledTimes(1);
});
