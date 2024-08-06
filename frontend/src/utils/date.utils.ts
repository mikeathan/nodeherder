export const toUnix = (date: Date) =>
  Math.floor(date.getTime() / 1000);

export const getDateRange = (
  hours: number,
): { from: Date; to: Date } => {
  const now = new Date();

  const date = new Date();
  date.setHours(hours);

  return { from: now, to: date };
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
