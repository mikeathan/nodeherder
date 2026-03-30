import { escapeHTML } from '@/utils/html';

describe('escapeHTML', () => {
  it('should escape HTML special characters', () => {
    const input = '<script>alert("XSS")</script> & "quotes"';
    const expected = '&lt;script&gt;alert(&quot;XSS&quot;)&lt;/script&gt; &amp; &quot;quotes&quot;';
    expect(escapeHTML(input)).toBe(expected);
  });

  it('should handle single quotes', () => {
    const input = "It's a test";
    const expected = 'It&#39;s a test';
    expect(escapeHTML(input)).toBe(expected);
  });

  it('should handle non-string inputs', () => {
    expect(escapeHTML(123)).toBe('123');
    expect(escapeHTML(true)).toBe('true');
  });

  it('should return empty string for null or undefined', () => {
    // @ts-ignore
    expect(escapeHTML(null)).toBe('');
    // @ts-ignore
    expect(escapeHTML(undefined)).toBe('');
  });
});
