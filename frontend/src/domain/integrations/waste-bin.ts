/** Fixed illustration colours keep provider text out of SVG/CSS values. */
const BIN_COLOURS = {
  black: { body: '#343e48', lid: '#242d35' },
  blue: { body: '#307fbd', lid: '#205d90' },
  brown: { body: '#95633e', lid: '#70472b' },
  green: { body: '#37875b', lid: '#256440' },
  grey: { body: '#7d8995', lid: '#596671' },
  red: { body: '#bd4d55', lid: '#91343d' },
  yellow: { body: '#d8b43d', lid: '#a58122' },
  orange: { body: '#d58a3e', lid: '#aa6227' },
  purple: { body: '#8e68ad', lid: '#684785' },
} as const;

function knownBinColour(colour: string | null): keyof typeof BIN_COLOURS | null {
  const key = colour?.trim().toLowerCase();
  if (key === 'gray') return 'grey';
  return key && Object.hasOwn(BIN_COLOURS, key) ? (key as keyof typeof BIN_COLOURS) : null;
}

/** Unknown colours stay neutral; the adjacent source text preserves the actual description. */
export function wasteBinColours(colour: string | null) {
  return BIN_COLOURS[knownBinColour(colour) ?? 'grey'];
}

/** Keep card headings short. Unknown colours retain the provider's service name instead. */
export function wasteBinTitle(colour: string | null): string | null {
  const key = knownBinColour(colour);
  return key ? `${key.charAt(0).toUpperCase()}${key.slice(1)} bin` : null;
}
