export const toUTC = (date: Date): Date => {
  // Create a new Date object to avoid modifying the original
  const utcDate = new Date(date.getTime());

  // Adjust for local time offset
  utcDate.setMinutes(
    utcDate.getMinutes() - utcDate.getTimezoneOffset(),
  );

  return utcDate;
};

export const utcToUnixTimestamp = (date: Date): number => {
  // Create a Date object from the UTC string
  // Convert to UTC milliseconds
  const utcMilliseconds = Date.UTC(
    date.getFullYear(),
    date.getMonth(),
    date.getDate(),
    date.getHours(),
    date.getMinutes(),
    date.getSeconds(),
  );
  const unixTimestamp = Math.floor(utcMilliseconds / 1000);
  return unixTimestamp;
};
export const toUnix = (date: Date): number =>
  Math.floor(date.getTime() / 1000);

export const getDateRange = (
  hours: number,
): { from: Date; to: Date } => {
  const now = new Date();

  const date = new Date();
  date.setHours(hours);

  return { from: date, to: now };
};

export const getWeekStartEndDate = (): {
  from: Date;
  to: Date;
} => {
  const today = new Date();

  // Get the day of the week (0-6, Sunday-Saturday)
  const dayOfWeek = today.getDay();

  // Calculate the start of the week (Monday)
  const weekStart = new Date(today);
  weekStart.setDate(today.getDate() - dayOfWeek + 1);

  // Calculate the end of the week (Sunday)
  const weekEnd = new Date(weekStart);
  weekEnd.setDate(weekEnd.getDate() + 6);

  return { from: weekStart, to: weekEnd };
};

export const getLastWeekStartEndDate = (): {
  from: Date;
  to: Date;
} => {
  const today = new Date();

  // Get the day of the week (0-6, Sunday-Saturday)
  const dayOfWeek = today.getDay();

  // Calculate the start of this week (Monday)
  const thisWeekStart = new Date(today);
  thisWeekStart.setDate(today.getDate() - dayOfWeek + 1);

  // Calculate the start of last week (Monday)
  const lastWeekStart = new Date(thisWeekStart);
  lastWeekStart.setDate(lastWeekStart.getDate() - 7);

  // Calculate the end of last week (Sunday)
  const lastWeekEnd = new Date(lastWeekStart);
  lastWeekEnd.setDate(lastWeekEnd.getDate() + 6);

  return { from: lastWeekStart, to: lastWeekEnd };
};

export const formatTimestamp = (
  timestamp: number,
): string => {
  const date = new Date(timestamp);
  return date.toLocaleString();
};
