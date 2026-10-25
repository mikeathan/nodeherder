import type { ConnectionValue, ProviderDefinition } from '@/types/integrations.type';
export function humanLabel(name: string): string {
  const words = name.replace(/_/g, ' ');
  return words.charAt(0).toUpperCase() + words.slice(1);
}
export function valueChoices(d: ProviderDefinition): { value: string; label: string }[] {
  const choices = Object.entries(d.inputs).map(([key, input]) => ({
    value: `input.${key}`,
    label: key === d.setup.select.input ? 'Address selected in Setup' : `${input.label} entered in Setup`,
  }));
  const seen = new Set<string>();
  Object.values(d.requests).forEach((r) =>
    Object.keys(r.captures ?? {}).forEach((key) => {
      if (!seen.has(key)) {
        seen.add(key);
        choices.push({
          value: `state.${key}`,
          label: `${key === 'property_token' ? 'Property reference' : key === 'lookup_id' ? 'Lookup reference' : humanLabel(key)} saved by the provider`,
        });
      }
    })
  );
  return choices;
}
export function valueSource(v: ConnectionValue): string {
  return typeof v === 'object' && v !== null ? v.ref : '';
}
export function fixedValue(text: string, kind: string): ConnectionValue {
  if (kind === 'boolean') return text === 'true';
  if (kind === 'null') return null;
  if (kind === 'number') {
    const value = Number(text);
    return Number.isFinite(value) ? value : text;
  }
  return text;
}

// A compact text comparison is enough for small provider files; no editor/diff runtime is needed.
export function providerChanges(current: string, shipped: string): string {
  const a = current.split('\n');
  const b = shipped.split('\n');
  const lines: string[] = [];
  for (let i = 0; i < Math.max(a.length, b.length); i++) {
    if (a[i] === b[i]) continue;
    lines.push(`Line ${i + 1}`);
    if (a[i] !== undefined) lines.push(`− ${a[i]}`);
    if (b[i] !== undefined) lines.push(`+ ${b[i]}`);
  }
  return lines.join('\n') || 'No differences.';
}
