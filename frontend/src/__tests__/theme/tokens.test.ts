import { readFileSync, readdirSync, statSync } from 'node:fs';
import { join, relative, resolve } from 'node:path';

/**
 * Guards that component styles only use custom properties the theme layer defines.
 *
 * This exists because one component shipped with four undefined names in its stylesheet —
 * `--border-subtle`, `--surface-raised`, `--accent-strong` and `--text-muted` — each with a
 * hardcoded hex fallback. Nothing failed: the fallback won on every theme, so a dark dashboard
 * showed a near-white 1px border and it looked like a deliberate design choice rather than a typo.
 * A fallback makes an undefined token invisible, which is exactly why it needs a test.
 *
 * Companion to `src/assets/styles/tokens.css`: that file owns the names, this asserts them.
 */

const SOURCE_ROOT = resolve(__dirname, '../../');
const STYLE_ROOT = resolve(SOURCE_ROOT, 'assets/styles');

/** Tokens declared anywhere in the style layer, including inside `:root` and theme blocks. */
function declaredTokens(): Set<string> {
  const declared = new Set<string>();
  for (const file of styleFiles(STYLE_ROOT)) {
    for (const match of readFileSync(file, 'utf8').matchAll(/(--[a-zA-Z0-9_-]+)\s*:/g)) {
      declared.add(match[1]);
    }
  }
  return declared;
}

/**
 * Tokens any component declares, wherever it declares them.
 *
 * A parent may set a property in its own scoped block for a child to read, which is legitimate and
 * inherits normally — `GroupDashboard` does exactly that for the heading height and gap. Excluding
 * those would make this guard report working code, and a guard that cries wolf gets deleted.
 */
function componentDeclaredTokens(): Set<string> {
  const declared = new Set<string>();
  for (const file of componentFiles(resolve(SOURCE_ROOT, 'components'))) {
    for (const match of readFileSync(file, 'utf8').matchAll(/(--[a-zA-Z0-9_-]+)\s*:/g)) {
      declared.add(match[1]);
    }
  }
  return declared;
}

function styleFiles(directory: string): string[] {
  const out: string[] = [];
  for (const entry of readdirSync(directory)) {
    const path = join(directory, entry);
    const stats = statSync(path);
    if (stats.isDirectory()) out.push(...styleFiles(path));
    else if (/\.(vue|css)$/.test(entry)) out.push(path);
  }
  return out;
}

function componentFiles(directory: string): string[] {
  const out: string[] = [];
  for (const entry of readdirSync(directory)) {
    const path = join(directory, entry);
    const stats = statSync(path);
    if (stats.isDirectory()) {
      if (entry === '__tests__') continue;
      out.push(...componentFiles(path));
    } else if (entry.endsWith('.vue')) {
      out.push(path);
    }
  }
  return out;
}

describe('theme tokens', () => {
  it('defines every custom property the stylesheets read', () => {
    const declared = declaredTokens();
    const missing = new Set<string>();
    for (const file of styleFiles(STYLE_ROOT)) {
      const source = readFileSync(file, 'utf8');
      for (const match of source.matchAll(/var\(\s*(--[a-zA-Z0-9_-]+)/g)) {
        if (!declared.has(match[1])) missing.add(match[1]);
      }
    }
    expect([...missing].sort()).toEqual([]);
  });

  /** Undefined custom properties read by components under one directory. */
  function undefinedTokensIn(directory: string): string[] {
    const defined = new Set([...declaredTokens(), ...componentDeclaredTokens()]);
    const missing: string[] = [];
    for (const file of componentFiles(resolve(SOURCE_ROOT, directory))) {
      for (const match of readFileSync(file, 'utf8').matchAll(/var\(\s*(--[a-zA-Z0-9_-]+)/g)) {
        // `--nh-kind-${kind}` is a template prefix, not a literal name.
        if (match[1].endsWith('-')) continue;
        if (!defined.has(match[1])) {
          missing.push(`${relative(SOURCE_ROOT, file)}: ${match[1]}`);
        }
      }
    }
    return [...new Set(missing)].sort();
  }

  /**
   * The regression this file was written for.
   *
   * Scoped to the Home dashboard, where the reported defect was: a tile styling with a name the
   * theme does not define renders its fallback on every theme, so the appearance is wrong in a way
   * no unit test and no type check can see.
   */
  it('defines every custom property the Home tiles read', () => {
    expect(undefinedTokensIn('components/home')).toEqual([]);
  });

  /**
   * The wireless viewer shipped with the same defect: a light `#f1f3f5` fallback painted a
   * near-white notice on a dark page. The fix is only durable if the names stay defined.
   */
  it('defines every custom property the wireless components read', () => {
    expect(undefinedTokensIn('components/wireless')).toEqual([]);
  });
});
