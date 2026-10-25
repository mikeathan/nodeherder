/** Reveal a field nested in collapsed form sections, without changing its keyboard tab order. */
export function revealConnectionField(id: string): HTMLElement | null {
  const field = document.getElementById(id);
  let parent = field?.parentElement;
  while (parent) {
    if (parent.tagName === 'DETAILS') (parent as HTMLDetailsElement).open = true;
    parent = parent.parentElement;
  }
  return field;
}

export function focusConnectionField(id: string) {
  const field = revealConnectionField(id);
  if (!field) return;
  // Container error targets need programmatic focus; actual inputs must stay in the tab sequence.
  if (!field.matches('input,select,textarea,button,a[href],[tabindex]')) field.tabIndex = -1;
  field.focus();
  field.scrollIntoView({ block: 'center' });
}
