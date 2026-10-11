/*
 * Promise-based confirmation for destructive or bridge-level actions (FE-03): screens call
 * `await confirm({...})` and the single ConfirmHost in the shell shows the dialog.
 */
import { shallowRef } from 'vue';

export type ConfirmRequest = {
  title: string;
  message: string;
  confirmLabel?: string;
  danger?: boolean;
};

type Pending = ConfirmRequest & { resolve: (ok: boolean) => void };

export const confirmState = shallowRef<Pending | null>(null);

export function confirm(request: ConfirmRequest): Promise<boolean> {
  confirmState.value?.resolve(false);
  return new Promise((resolve) => {
    confirmState.value = { ...request, resolve };
  });
}

export function settleConfirm(ok: boolean) {
  const current = confirmState.value;
  confirmState.value = null;
  current?.resolve(ok);
}
