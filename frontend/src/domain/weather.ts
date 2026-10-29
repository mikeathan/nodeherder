/** Present normalized conditions as readable text; unrecognized or absent data remains explicit. */
export function weatherLabel(value: string | undefined): string {
  return value ? value.replaceAll('_', ' ').replace(/^./, (c) => c.toUpperCase()) : 'Unavailable';
}
