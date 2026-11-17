/**
 * Whether the device-backed wireless observation group is hidden from the board.
 *
 * Stored in browser storage rather than with the board layout, and deliberately so. The board's
 * order and visibility are persisted server-side against *instance* ids — every entry there refers
 * to something the operator configured. Wireless observation has no instance: it appears because a
 * device is enrolled, and disappears when it is not. Inventing a synthetic board entry would mean a
 * schema change and a validator that accepts ids which resolve to nothing.
 *
 * Hiding it is a presentation preference, exactly the class of state `utils/storage` is documented
 * to hold (FE-05): it never leaves this browser, and losing it costs nothing but one click.
 */
import { ref } from 'vue';
import { safeStorage } from '@/utils/storage';

const STORAGE_KEY = 'nodeherder.integrations.hideObservation';

/** Read the stored preference, defaulting to shown. */
function read(): boolean {
  return safeStorage.getItem(STORAGE_KEY) === 'true';
}

/**
 * Shared across the page and the board.
 *
 * Module-level rather than per-call, because the page renders both the toggle state and the board
 * that consumes it. One instance of the source of truth.
 */
const hidden = ref(read());

/** Whether the observation group is currently hidden. */
export function observationHidden() {
  return hidden;
}

/**
 * Hide or reveal the observation group.
 *
 * The write happens here rather than in a `watch`. A module-level watcher is not attached to a
 * component scope, so its callback never ran and the preference silently failed to persist — the
 * group came back on the next reload with nothing to explain why. Writing in the setter makes the
 * two impossible to separate.
 */
export function setObservationHidden(value: boolean): void {
  hidden.value = value;
  safeStorage.setItem(STORAGE_KEY, String(value));
}
