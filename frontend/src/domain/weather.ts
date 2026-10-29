/** Format optional normalized weather quantities without substituting zero for gaps. */
export function weatherValue(value: number | null | undefined, unit: string): string {
  return value == null ? 'Unavailable' : `${Math.round(value)}${unit}`;
}
/** Present normalized conditions as readable text; unrecognized or absent data remains explicit. */
export function weatherLabel(value: string | undefined): string {
  return value ? value.replaceAll('_', ' ').replace(/^./, (c) => c.toUpperCase()) : 'Unavailable';
}
