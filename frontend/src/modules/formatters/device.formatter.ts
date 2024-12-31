import {
  mdiPowerPlug,
  mdiBattery,
  mdiBatteryCharging,
  mdiAlertCircleOutline,
} from '@mdi/js';

export function getPowerSourceIcon(
  power_source: string,
  value: number | undefined
): { icon: string | null; color: string | null } {
  if (!power_source) {
    return { icon: null, color: null };
  }

  if (power_source.toLowerCase().includes('mains')) {
    return { icon: mdiPowerPlug, color: null };
  }

  if (value === undefined) {
    return { icon: mdiAlertCircleOutline, color: 'orange' }; // Indicate unknown with orange alert
  }

  let icon = mdiBattery;
  let color: string | null = null;

  if (value >= 95) {
    icon = mdiBattery;
    color = 'green';
  } else if (value >= 70) {
    icon = mdiBattery;
    color = 'lightgreen';
  } else if (value >= 40) {
    icon = mdiBattery;
    color = 'orange';
  } else if (value >= 15) {
    icon = mdiBattery;
    color = 'darkorange';
  } else {
    icon = mdiBattery;
    color = 'red'; // Low battery, use red
  }

  return { icon, color };
}
