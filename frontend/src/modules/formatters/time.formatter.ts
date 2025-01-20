import { Nullable } from '@/types/types.type';

export function toMillisecs(minutes: number): number {
  return minutes * 60000;
}

export function toMinutes(
  millisecs: Nullable<number>
): Nullable<number> {
  if (millisecs == null) {
    return null;
  }
  if (millisecs >= 1000) {
    return (millisecs /= 60000);
  }
  return millisecs;
}
