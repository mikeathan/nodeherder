/**
 * Safely escapes HTML special characters to prevent XSS.
 * @param str The string to escape.
 * @returns The escaped string.
 */
export function escapeHTML(str: string | number | boolean): string {
  if (str === null || str === undefined) {
    return '';
  }
  const s = String(str);
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}
