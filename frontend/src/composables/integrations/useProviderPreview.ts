import { onBeforeUnmount, ref, shallowRef, type Ref } from 'vue';
import { confirm } from '@/composables/useConfirm';
import type { ConnectionScope } from '@/composables/useConnections';
import type { ConnectionSetup, DefinitionValidation, DestinationReview } from '@/types/integrations.type';

/** Owns ephemeral sample/preview data; tests cannot replace saved schedules or provider files. */
export function useProviderPreview(
  scope: ConnectionScope,
  key: Ref<string>,
  validate: () => Promise<DefinitionValidation | undefined>,
  review: () => DestinationReview | undefined
) {
  const testInstance = ref('');
  const sample = ref('');
  const preview = shallowRef<ConnectionSetup | null>(null);
  onBeforeUnmount(() => {
    sample.value = '';
    preview.value = null;
  });
  async function test(mode: 'offline' | 'live') {
    const v = await validate();
    if (!v) return;
    let input: unknown;
    if (mode === 'offline') {
      try {
        input = JSON.parse(sample.value);
      } catch {
        scope.error.value = { code: 'invalid_input', message: 'Enter a valid JSON sample.' };
        return;
      }
    }
    if (
      mode === 'live' &&
      !(await confirm({
        title: 'Send a test request?',
        message:
          'This sends a saved-schedule request using the selected connection’s private inputs and token. The preview does not replace its saved results.',
        confirmLabel: 'Send test request',
      }))
    )
      return;
    const result = await scope.perform(() =>
      scope.api.test(key.value, v.yaml, testInstance.value, mode, input, review())
    );
    if (result) {
      preview.value = result;
      if (result.operation_id)
        scope.observe(result.operation_id, async (signal) => {
          const updated = await scope.api.setup(result.id, signal);
          if (scope.isActive() && !signal.aborted) preview.value = updated;
          return updated;
        });
    }
  }

  return { testInstance, sample, preview, test };
}
