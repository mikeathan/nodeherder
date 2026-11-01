/** Time/random suffix used for non-UUID local identifiers. */
export function genenerateUniqueId(): string {
  const dateStr = Date.now().toString(36); // convert num to base 36 and stringify

  const randomStr = Math.random()
    .toString(36)
    .substring(2, 8); // start at index 2 to skip decimal point

  return `${dateStr}-${randomStr}`;
}

/** Formats 16 bytes as a canonical version-4 UUID with RFC 4122 version/variant bits. */
export function uuidFromBytes(bytes: Uint8Array): string {
  const value = new Uint8Array(bytes);
  value[6] = (value[6] & 0x0f) | 0x40;
  value[8] = (value[8] & 0x3f) | 0x80;
  const hex = Array.from(value, (byte) => byte.toString(16).padStart(2, '0')).join('');
  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`;
}

/**
 * Canonical random UUID that works in non-secure HTTP contexts, where Firefox
 * does not expose `crypto.randomUUID`. Uses `getRandomValues` when available and
 * falls back to Math.random only for local identifiers when Web Crypto is absent.
 */
export function createUUID(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID();
  }
  const bytes = new Uint8Array(16);
  if (typeof crypto !== 'undefined' && typeof crypto.getRandomValues === 'function') {
    crypto.getRandomValues(bytes);
  } else {
    for (let index = 0; index < bytes.length; index += 1) bytes[index] = Math.floor(Math.random() * 256);
  }
  return uuidFromBytes(bytes);
}

/** Backward-compatible conversation identifier; all UUID creation shares `createUUID`. */
export function generateConversationId(): string {
  return createUUID();
}
