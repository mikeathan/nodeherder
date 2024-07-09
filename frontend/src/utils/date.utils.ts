export const toUnix = (date: Date) =>
  Math.floor(date.getTime() / 1000);
