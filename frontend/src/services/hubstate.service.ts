import { AppConfig } from '@/types/settings.type';
import { Device } from '@/types/device';
const baseUrl = import.meta.env.VITE_API_BASE_URL;

export async function fetchHubState(): Promise<{ devices: Device[]; config: AppConfig }> {
  const res = await fetch(`${baseUrl}/api/hubstate`);
  if (!res.ok) {
    console.error('[DEBUG]hubstate fetch failed', res);
    throw new Error('hubstate fetch failed');
  }

  return await res.json();
}
