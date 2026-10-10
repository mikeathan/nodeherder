/*
 * Panel mode layout (spec 007 US-07, FR-10). Turns a Home area into a climate dial (first
 * temperature, with humidity when present) and one large button per device: devices with an
 * on/off state get a single toggle button, other values get one button each.
 */
import { DashboardGroup } from '@/types/settings.type';
import { Device } from '@/types/device';
import { groupEntities, GroupEntity } from './dashboard';
import { exposeLabel, stateExposeOf } from './exposes';

export type PanelButton = GroupEntity & { label: string };
export type PanelRoom = { name: string; climate: GroupEntity | null; humidity: GroupEntity | null; buttons: PanelButton[] };

const TEMPERATURE = new Set(['temperature', 'local_temperature']);

/** Device name without the room name in front ("Kitchen light" in "Kitchen" → "Light"). */
export function shortName(deviceName: string, roomName: string): string {
  const name = deviceName.trim();
  const prefix = roomName.trim().toLowerCase();
  if (prefix && name.toLowerCase().startsWith(prefix + ' ') && name.length > prefix.length + 1) {
    const rest = name.slice(prefix.length + 1);
    return rest.charAt(0).toUpperCase() + rest.slice(1);
  }
  return name;
}

export function panelRoom(group: DashboardGroup, lookup: (id: string) => Device | undefined): PanelRoom {
  const items = groupEntities(group);
  const climate = items.find((i) => TEMPERATURE.has(i.expose)) ?? null;
  const humidity = items.find((i) => i.expose === 'humidity' && (!climate || i.deviceId === climate.deviceId)) ?? items.find((i) => i.expose === 'humidity') ?? null;
  const toggled = new Set<string>();
  const buttons: PanelButton[] = [];
  for (const item of items) {
    if (item === climate || item === humidity) continue;
    const device = lookup(item.deviceId);
    if (device && stateExposeOf(device)) {
      if (toggled.has(item.deviceId)) continue;
      toggled.add(item.deviceId);
      buttons.push({ ...item, label: shortName(device.friendly_name, group.name) });
    } else {
      buttons.push({ ...item, label: exposeLabel(item.expose) });
    }
  }
  return { name: group.name, climate, humidity, buttons };
}
