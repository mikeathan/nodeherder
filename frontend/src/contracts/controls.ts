import { TimePicker } from '@/types/controls.type';

export function toTimePicker(time: string): TimePicker {
  if (!time || time === '') {
    return { hours: 0, minutes: 0 };
  }
  const [hours, minutes] = time.split(':').map(Number);
  return { hours, minutes };
}
