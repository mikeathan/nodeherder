import { credentialLabel, credentialSlots } from '@/domain/integrations/credentials';
import type { ProviderAuth } from '@/types/integrations';

test('declares OAuth slots and labels them independently of saved values', () => {
  const auth = {
    mode: 'oauth2_client_credentials',
    client_id_slot: 'client_id',
    client_secret_slot: 'client_secret',
  } as ProviderAuth;
  expect(credentialSlots(auth)).toEqual(['client_id', 'client_secret']);
  expect(credentialLabel(auth, 'client_id')).toBe('OAuth client ID');
  expect(credentialLabel(auth, 'client_secret')).toBe('OAuth client secret');
});

test('returns no slots for anonymous authentication and one for key modes', () => {
  expect(credentialSlots({ mode: 'none' } as ProviderAuth)).toEqual([]);
  expect(credentialSlots({ mode: 'bearer', secret_slot: 'token' } as ProviderAuth)).toEqual(['token']);
  expect(credentialLabel({ mode: 'bearer', secret_slot: 'token' } as ProviderAuth, 'token')).toBe('API key / token');
});
