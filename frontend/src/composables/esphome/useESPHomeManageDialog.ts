import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { confirm } from '@/composables/useConfirm';
import { useHub } from '@/composables/useHub';
import { useESPHomeEnrollmentList } from '@/composables/esphome/useESPHomeEnrollmentList';
import { removeESPHomeDevice, updateESPHomeDevice } from '@/services/esphome';
import type { ESPHomeEnrollment, ESPHomeEnrollmentUpdate } from '@/contracts/esphome';
import { ConnectionTypes } from '@/types/device.type';
import {
  ESPHOME_EDIT_INPUT_ERROR,
  ESPHOME_FORM_DEFAULT_PORT,
  ESPHOME_REMOVE_CONFIRMATION,
  ESPHOME_REMOVE_RETRY_FAILED,
  ESPHOME_SAVE_RETRY_FAILED,
  createEnrollmentUpdate,
  diagnosticMessage as diagnosticText,
  enrollmentDisplayName,
  validESPHomePort,
  type ESPHomeKeyMode,
} from '@/domain/esphome';

/**
 * Owns the ESPHome management dialog's polling, edit drafts and fenced
 * mutations. The component only renders state and forwards user actions.
 */
export function useESPHomeManageDialog(open: () => boolean) {
  const enrollmentList = useESPHomeEnrollmentList();
  const { items, loading, loaded, loadError, refresh, cancelLoad } = enrollmentList;
  const { devices, initialized } = useHub();

  const unmatchedDevices = computed(() => {
    const enrolled = new Set(items.value.map((item) => item.id));
    return devices.value.filter(
      (device) =>
        (device.provider === ConnectionTypes.ESPHome || device.connection_type === ConnectionTypes.ESPHome) &&
        !enrolled.has(device.id)
    );
  });

  const error = ref('');
  // Closing retires outstanding mutation completions without retrying server writes.
  let generation = 0;
  const removing = ref('');
  const editing = ref('');
  const editName = ref('');
  const editAddress = ref('');
  const editPort = ref(ESPHOME_FORM_DEFAULT_PORT);
  const keyMode = ref<ESPHomeKeyMode>('keep');
  const editKey = ref('');
  const saving = ref(false);
  const changing = computed(() => saving.value || !!removing.value);

  /** Name shown for a saved enrollment, including registry fallback. */
  function displayName(item: ESPHomeEnrollment): string {
    return enrollmentDisplayName(item, devices.value);
  }

  function stop() {
    generation++;
    enrollmentList.stop();
  }

  function closeEdit() {
    editing.value = '';
    editKey.value = '';
  }

  function resetDialog() {
    enrollmentList.reset();
    closeEdit();
    error.value = '';
    saving.value = false;
    removing.value = '';
  }

  watch(open, (isOpen) => {
    if (isOpen) {
      enrollmentList.start(() => !changing.value);
      return;
    }
    stop();
    resetDialog();
  }, { immediate: true });
  onBeforeUnmount(stop);

  /** Removes one enrollment after confirmation; stale completions are ignored. */
  async function remove(item: ESPHomeEnrollment) {
    if (changing.value) return;
    const owner = generation;
    const approved = await confirm({
      title: `Remove ${displayName(item)}?`,
      ...ESPHOME_REMOVE_CONFIRMATION,
      danger: true,
    });
    if (!approved || owner !== generation || !open() || changing.value) return;
    cancelLoad();
    removing.value = item.id;
    error.value = '';
    try {
      await removeESPHomeDevice(item.id);
      if (owner === generation) {
        if (editing.value === item.id) closeEdit();
        await refresh();
      }
    } catch {
      if (owner === generation) error.value = ESPHOME_REMOVE_RETRY_FAILED;
    } finally {
      if (owner === generation) removing.value = '';
    }
  }

  /** Starts editing a saved enrollment with a preserved-key default. */
  function edit(item: ESPHomeEnrollment) {
    if (changing.value) return;
    editing.value = item.id;
    editName.value = item.name || '';
    editAddress.value = item.address;
    editPort.value = item.port;
    keyMode.value = 'keep';
    editKey.value = '';
    error.value = '';
  }

  function connectionInput(): ESPHomeEnrollmentUpdate {
    if (!editAddress.value.trim() || !validESPHomePort(editPort.value) || (keyMode.value === 'replace' && !editKey.value.trim())) {
      throw new Error(ESPHOME_EDIT_INPUT_ERROR);
    }
    return createEnrollmentUpdate(editName.value, editAddress.value, editPort.value, keyMode.value, editKey.value);
  }

  /** Saves the connection replacement and refreshes the redacted list. */
  async function save() {
    if (changing.value) return;
    const owner = generation;
    let input: ESPHomeEnrollmentUpdate;
    try {
      input = connectionInput();
    } catch (cause) {
      error.value = cause instanceof Error ? cause.message : ESPHOME_EDIT_INPUT_ERROR;
      return;
    }
    cancelLoad();
    saving.value = true;
    error.value = '';
    try {
      await updateESPHomeDevice(editing.value, input);
      // A completed server mutation must not overwrite a reopened dialog's draft.
      if (owner !== generation) return;
      closeEdit();
      await refresh();
    } catch (cause) {
      if (owner === generation) error.value = cause instanceof Error ? cause.message : ESPHOME_SAVE_RETRY_FAILED;
    } finally {
      if (owner === generation) saving.value = false;
    }
  }

  /** Classifies a connection diagnostic for display. */
  function diagnosticMessage(code?: string): string {
    return diagnosticText(code);
  }

  return {
    items,
    loading,
    loaded,
    initialized,
    loadError,
    error,
    removing,
    editing,
    editName,
    editAddress,
    editPort,
    keyMode,
    editKey,
    saving,
    changing,
    unmatchedDevices,
    displayName,
    remove,
    edit,
    closeEdit,
    save,
    diagnosticMessage,
  };
}
