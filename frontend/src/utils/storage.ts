/*
 * Browser storage that never throws (private windows, blocked site data, SSR/tests).
 * Only presentation preferences are stored (FE-05); household data never is.
 */
export type SafeStorage = Pick<Storage, 'getItem' | 'setItem' | 'removeItem'>;

function resolve(): Storage | null {
  try {
    return typeof window !== 'undefined' && window.localStorage ? window.localStorage : null;
  } catch {
    return null;
  }
}

export const safeStorage: SafeStorage = {
  getItem(key) {
    try {
      return resolve()?.getItem(key) ?? null;
    } catch {
      return null;
    }
  },
  setItem(key, value) {
    try {
      resolve()?.setItem(key, value);
    } catch {
      /* storage full or blocked: the preference lasts for this session only */
    }
  },
  removeItem(key) {
    try {
      resolve()?.removeItem(key);
    } catch {
      /* ignore */
    }
  },
};
