export const toUTC = (date: Date): Date => {
  // Create a new Date object to avoid modifying the original
  const utcDate = new Date(date.getTime());

  // Adjust for local time offset
  utcDate.setMinutes(utcDate.getMinutes() - utcDate.getTimezoneOffset());

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
    date.getSeconds()
  );
  const unixTimestamp = Math.floor(utcMilliseconds / 1000);
  return unixTimestamp;
};
export const toUnix = (date: Date): number => Math.floor(date.getTime() / 1000);

export const getDateRange = (hours: number): { from: Date; to: Date } => {
  const now = new Date();
  const from = new Date(now.getTime() + hours * 60 * 60 * 1000);
  return { from, to: now };
};
// Returns the range from local midnight today to now
export const getTodayRange = (): { from: Date; to: Date } => {
  const now = new Date();
  const midnight = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 0, 0, 0, 0);
  return { from: midnight, to: now };
};

// Returns the range for yesterday: local midnight of previous day to local midnight of today
export const getYesterdayRange = (): { from: Date; to: Date } => {
  const now = new Date();
  const todayMidnight = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 0, 0, 0, 0);
  const yesterdayMidnight = new Date(todayMidnight.getTime() - 24 * 60 * 60 * 1000);
  return { from: yesterdayMidnight, to: todayMidnight };
};

export const getWeekStartEndDate = (): {
  from: Date;
  to: Date;
} => {
  const today = new Date();
  const dayOfWeek = today.getDay();
  const weekStart = new Date(today);
  weekStart.setDate(today.getDate() - dayOfWeek + 1);
  // normalize to local midnight
  const from = new Date(weekStart.getFullYear(), weekStart.getMonth(), weekStart.getDate(), 0, 0, 0, 0);
  // end of week = Sunday 23:59:59.999
  const weekEnd = new Date(from);
  weekEnd.setDate(weekEnd.getDate() + 6);
  const to = new Date(weekEnd.getFullYear(), weekEnd.getMonth(), weekEnd.getDate(), 23, 59, 59, 999);
  return { from, to };
};

export const getLastWeekStartEndDate = (): {
  from: Date;
  to: Date;
} => {
  const today = new Date();
  const dayOfWeek = today.getDay();
  const thisWeekStart = new Date(today);
  thisWeekStart.setDate(today.getDate() - dayOfWeek + 1);
  const lastWeekStart = new Date(thisWeekStart);
  lastWeekStart.setDate(lastWeekStart.getDate() - 7);
  const from = new Date(lastWeekStart.getFullYear(), lastWeekStart.getMonth(), lastWeekStart.getDate(), 0, 0, 0, 0);
  const lastWeekEnd = new Date(from);
  lastWeekEnd.setDate(lastWeekEnd.getDate() + 6);
  const to = new Date(lastWeekEnd.getFullYear(), lastWeekEnd.getMonth(), lastWeekEnd.getDate(), 23, 59, 59, 999);
  return { from, to };
};

export const formatTimestamp = (timestamp: number): string => {
  try {
    const date = new Date(timestamp);
    return date.toISOString();
  } catch (error) {
    return 'Invalid Date';
  }
};

export function formatDuration(durationMs: number): string {
  const minutes = Math.floor(durationMs / 60000);
  const hours = Math.floor(minutes / 60);
  const remainingMins = minutes % 60;

  if (hours > 0) {
    return `${hours}h ${remainingMins}m`;
  }
  return `${minutes}m`;
}

export function formatTime(timestamp: number): string {
  return new Date(timestamp).toLocaleTimeString('en-GB', {
    hour: '2-digit',
    minute: '2-digit',
  });
}

export function parseTimestamp(value: string | number | undefined, fallback: number): number {
  if (!value) return fallback;
  return typeof value === 'string' ? new Date(value).getTime() : value;
}

// formatting: Today / Yesterday / same-year / cross-year.
export function formatSmartDate(dateInput: Date | number): string {
  const date = typeof dateInput === 'number' ? new Date(dateInput) : dateInput;
  const now = new Date();
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  const yesterday = new Date(today);
  yesterday.setDate(yesterday.getDate() - 1);

  const dateOnly = new Date(date.getFullYear(), date.getMonth(), date.getDate());

  const timeStr = date.toLocaleTimeString('en-GB', {
    hour: '2-digit',
    minute: '2-digit',
  });

  if (dateOnly.getTime() === today.getTime()) {
    return `Today ${timeStr}`;
  }

  if (dateOnly.getTime() === yesterday.getTime()) {
    return `Yesterday ${timeStr}`;
  }

  if (date.getFullYear() === now.getFullYear()) {
    const dateStr = date.toLocaleDateString('en-GB', {
      day: '2-digit',
      month: 'short',
    });
    return `${dateStr} ${timeStr}`;
  }

  const dateStr = date.toLocaleDateString('en-GB', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  });
  return `${dateStr} ${timeStr}`;
}
