import { IconProps } from '@/types/icon.type';
import {
  mdiPowerPlug,
  mdiBattery,
  mdiBatteryCharging,
  mdiAlertCircleOutline,
  mdiSignal,
  mdiBattery10,
  mdiBattery20,
  mdiBattery40,
  mdiBattery70,
} from '@mdi/js';

export function getSignalIcon(value: number): IconProps {
  return {
    name: mdiSignal,
    color: 'white',
    tooltip: `${value} LQI`,
  };
}

export function getPowerSourceIcon(
  power_source: string,
  value: number | undefined
): IconProps {
  const tooltip = `${power_source} ${value}`;
  if (!power_source) {
    return { name: '', color: '', tooltip: tooltip };
  }

  if (power_source.toLowerCase().includes('mains')) {
    return {
      name: mdiPowerPlug,
      color: 'gray',
      tooltip: tooltip,
    };
  }

  if (value === undefined) {
    return {
      name: mdiAlertCircleOutline,
      color: 'orange',
      tooltip: tooltip,
    };
  }

  let name = mdiBattery;
  let color: string | null = null;

  if (value >= 95) {
    name = mdiBattery;
    color = 'green';
  } else if (value >= 70) {
    name = mdiBattery70;
    color = 'lightgreen';
  } else if (value >= 40) {
    name = mdiBattery40;
    color = 'orange';
  } else if (value >= 15) {
    name = mdiBattery20;
    color = 'darkorange';
  } else {
    name = mdiBattery10;
    color = 'red';
  }

  return { name, color, tooltip: tooltip };
}
