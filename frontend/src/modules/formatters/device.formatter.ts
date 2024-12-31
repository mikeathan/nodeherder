import { IconProps } from '@/types/icon.type';
import {
  mdiPowerPlug,
  mdiBattery,
  mdiBatteryCharging,
  mdiAlertCircleOutline,
} from '@mdi/js';

export function getPowerSourceIcon(
  power_source: string,
  value: number | undefined
): IconProps {
  if (!power_source) {
    return { name: '', color: '' };
  }

  if (power_source.toLowerCase().includes('mains')) {
    return { name: mdiPowerPlug, color: 'gray' };
  }

  if (value === undefined) {
    return { name: mdiAlertCircleOutline, color: 'orange' }; // Indicate unknown with orange alert
  }

  let name = mdiBattery;
  let color: string | null = null;

  if (value >= 95) {
    name = mdiBattery;
    color = 'green';
  } else if (value >= 70) {
    name = mdiBattery;
    color = 'lightgreen';
  } else if (value >= 40) {
    name = mdiBattery;
    color = 'orange';
  } else if (value >= 15) {
    name = mdiBattery;
    color = 'darkorange';
  } else {
    name = mdiBattery;
    color = 'red'; // Low battery, use red
  }

  return { name, color };
}
