import { AppConfig } from '@/types/settings.type';
import { Device } from '@/types/device';

export async function fetchHubState(): Promise<{ devices: Device[]; config: AppConfig }> {
  const res = await fetch('/hubstate');
  if (!res.ok) throw new Error('hubstate fetch failed');
  return await res.json();
}
