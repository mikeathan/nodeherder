export function getNowTimeString(): string {
  const now = new Date();
  now.setMinutes(0, 0, 0);
  return toHourMinuteString(now);
}
export function getNowTime(): Date {
  const now = new Date();
  now.setMinutes(0, 0, 0);
  return now;
}

export function convertTimeToDate(time: string): Date {
  if (!time || time === '') {
    return new Date();
  }
  const [hours, minutes] = time.split(':').map(Number);

  const date = new Date();
  date.setHours(hours);
  date.setMinutes(minutes);
  date.setSeconds(0);
  date.setMilliseconds(0);
  return date;
}

export function toHourMinuteString(date: Date): string {
  const hours = date.getHours().toString().padStart(2, '0');
  const minutes = date.getMinutes().toString().padStart(2, '0');
  return `${hours}:${minutes}`;
}
