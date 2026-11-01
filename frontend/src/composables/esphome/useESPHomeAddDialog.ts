import { onBeforeUnmount, ref } from 'vue';
import { store } from '@/store';
import { ESPHOME_MESSAGES, type ESPHomeCandidate } from '@/contracts/esphome';
import { addESPHomeDevice, discoverESPHomeDevices } from '@/services/esphome';
import {
  ESPHOME_ADDED_NOTICE,
  ESPHOME_ADD_INPUT_ERROR,
  ESPHOME_DISCOVERY_FALLBACK,
  ESPHOME_FORM_DEFAULT_PORT,
  createEnrollmentInput,
  validESPHomeInput,
} from '@/domain/esphome';

/**
 * Owns the add-dialog form, discovery candidates and cancellable requests.
 * The dialog component only renders state and emits close/added events.
 */
export function useESPHomeAddDialog(hooks: { onClose: () => void; onAdded: () => void }) {
  const name = ref('');
  const address = ref('');
  const port = ref(ESPHOME_FORM_DEFAULT_PORT);
  const key = ref('');
  const busy = ref(false);
  const scanning = ref(false);
  const scanned = ref(false);
  const candidates = ref<ESPHomeCandidate[]>([]);
  const error = ref('');
  let request: AbortController | undefined;

  function resetForm() {
    name.value = '';
    address.value = '';
    port.value = ESPHOME_FORM_DEFAULT_PORT;
    key.value = '';
    error.value = '';
    scanned.value = false;
    candidates.value = [];
  }

  /** Cancels a pending request, discards the draft and closes the dialog. */
  function close() {
    if (busy.value) return;
    request?.abort();
    resetForm();
    hooks.onClose();
  }

  /** Runs one bounded discovery scan; selection fills the form without enrolling. */
  async function discover() {
    if (busy.value || scanning.value) return;
    scanning.value = true;
    error.value = '';
    request = new AbortController();
    try {
      candidates.value = await discoverESPHomeDevices(request.signal);
      scanned.value = true;
    } catch (cause) {
      if (!(cause instanceof Error && cause.name === 'AbortError')) error.value = ESPHOME_DISCOVERY_FALLBACK;
    } finally {
      scanning.value = false;
      request = undefined;
    }
  }

  /** Copies an untrusted candidate into editable fields. */
  function selectCandidate(candidate: ESPHomeCandidate) {
    name.value = candidate.name;
    address.value = candidate.address;
    port.value = candidate.port;
  }

  /** Persists the enrollment; a successful response means enrolled, not online. */
  async function submit() {
    if (busy.value || scanning.value) return;
    if (!validESPHomeInput(address.value, port.value)) {
      error.value = ESPHOME_ADD_INPUT_ERROR;
      return;
    }
    busy.value = true;
    error.value = '';
    request = new AbortController();
    try {
      await addESPHomeDevice(createEnrollmentInput(name.value, address.value, port.value, key.value), request.signal);
      busy.value = false;
      close();
      void store.dispatch('alerts/showSuccess', ESPHOME_ADDED_NOTICE);
      hooks.onAdded();
    } catch (cause) {
      if (cause instanceof Error && cause.name === 'AbortError') return;
      error.value = cause instanceof Error ? cause.message : ESPHOME_MESSAGES.unreachable;
    } finally {
      busy.value = false;
      request = undefined;
    }
  }

  onBeforeUnmount(() => request?.abort());

  return { name, address, port, key, busy, scanning, scanned, candidates, error, close, discover, selectCandidate, submit };
}
