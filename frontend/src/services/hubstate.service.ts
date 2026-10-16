import type { HubStateSnapshot } from '@/types/hub.type';
import { get } from '@/contracts/api';

export async function fetchHubState(signal?: AbortSignal): Promise<HubStateSnapshot> {
  const res = await get(`hubstate`, { signal });
  if (!res.ok) {
    console.error('hubstate fetch failed', res);
    throw new Error('hubstate fetch failed ');
  }
  return await res.json();
}
