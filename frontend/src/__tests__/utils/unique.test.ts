import { createUUID, generateConversationId, uuidFromBytes } from '@/utils/unique';

const v4Pattern = /^[\da-f]{8}-[\da-f]{4}-4[\da-f]{3}-[89ab][\da-f]{3}-[\da-f]{12}$/;

test('formats canonical version-4 UUIDs from bytes', () => {
  const bytes = Uint8Array.from({ length: 16 }, (_, index) => index);
  expect(uuidFromBytes(bytes)).toBe('00010203-0405-4607-8809-0a0b0c0d0e0f');
});

test('creates UUIDs in insecure-context-safe canonical form', () => {
  expect(createUUID()).toMatch(v4Pattern);
  expect(generateConversationId()).toMatch(v4Pattern);
  expect(createUUID()).not.toBe(createUUID());
});

test('falls back when crypto.randomUUID is unavailable', () => {
  const descriptor = Object.getOwnPropertyDescriptor(globalThis.crypto, 'randomUUID');
  Object.defineProperty(globalThis.crypto, 'randomUUID', { configurable: true, value: undefined });
  try {
    expect(createUUID()).toMatch(v4Pattern);
  } finally {
    if (descriptor) Object.defineProperty(globalThis.crypto, 'randomUUID', descriptor);
  }
});
