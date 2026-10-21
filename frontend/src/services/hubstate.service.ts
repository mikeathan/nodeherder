import type { HubStateSnapshot } from '@/types/hub.type';
import { get } from '@/contracts/api';
import { isHubStateSnapshot } from '@/contracts/hub';

export async function fetchHubState(signal?: AbortSignal): Promise<HubStateSnapshot> {
  const res = await get(`hubstate`, { signal });
  if (!res.ok) {
    console.error('hubstate fetch failed', res);
    throw new Error('hubstate fetch failed ');
  }
  const data: unknown = await res.json();
  if (!isHubStateSnapshot(data)) throw new Error('Invalid hub snapshot');
  return data;
}
