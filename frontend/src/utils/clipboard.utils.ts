export async function copyToClipboard(text: string): Promise<void> {
  try {
    if (navigator.clipboard && window.isSecureContext) {
      // Modern approach (HTTPS or localhost)
      await navigator.clipboard.writeText(text);
    } else {
      // Fallback for non-secure HTTP contexts
      const textArea = document.createElement('textarea');
      textArea.value = text;
      textArea.style.position = 'absolute';
      textArea.style.left = '-999999px';
      document.body.prepend(textArea);
      textArea.select();
      document.execCommand('copy');
      textArea.remove();
    }
  } catch (err) {
    console.error('Failed to copy text: ', err);
    throw err;
  }
}
