import type { Devices } from './device';
import type { AppConfig } from './settings.type';

export type HubLoadStatus = 'idle' | 'loading' | 'ready' | 'failed';

export type HubStateSnapshot = {
  devices: Devices;
  config: AppConfig;
};
