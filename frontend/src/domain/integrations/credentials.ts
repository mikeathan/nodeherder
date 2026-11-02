import type { ProviderAuth } from '@/types/integrations';

/** Declared private slots for the active auth mode; values are never derived from saved data. */
export function credentialSlots(auth: ProviderAuth | null | undefined): string[] {
  if (!auth || auth.mode === 'none') return [];
  if (auth.mode === 'oauth2_client_credentials') {
    return [auth.client_id_slot, auth.client_secret_slot].filter((slot): slot is string => Boolean(slot));
  }
  return auth.secret_slot ? [auth.secret_slot] : [];
}

/** Human label for one declared credential slot. */
export function credentialLabel(auth: ProviderAuth | null | undefined, slot: string): string {
  if (auth?.client_id_slot === slot) return 'OAuth client ID';
  if (auth?.client_secret_slot === slot) return 'OAuth client secret';
  return 'API key / token';
}
