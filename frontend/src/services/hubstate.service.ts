import { AppConfig } from '@/types/settings.type';
import { Device } from '@/types/device';
import { get } from '@/contracts/api';

export async function fetchHubState(): Promise<{ devices: Device[]; config: AppConfig }> {
  const res = await get(`hubstate`);
  if (!res.ok) {
    console.error('hubstate fetch failed', res);
    throw new Error('hubstate fetch failed ');
  }
  return await res.json();
}
