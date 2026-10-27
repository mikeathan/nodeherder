import type { HomeLayout } from '@/types/home.type';
import { parseHomeLayout, parseLegacyHomeGroups } from './home';
import { importLegacyHome } from '@/domain/home-compatibility';
import { cloneHome } from '@/domain/home-layout';

/** Versioned reference-only file format; independent from provider/resource payload versions. */
export const HOME_EXPORT_FORMAT = 'nodeherder.home/v1';

/** Decode a Home or legacy area export into a validated proposal against the captured revision.
 * Imported revision metadata is never trusted as permission to overwrite newer accepted state. */
export function parseHomeExport(value: unknown, current: HomeLayout): HomeLayout {
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    throw new Error('Choose a NodeHerder Home or areas export.');
  }
  const raw = value as Record<string, unknown>;
  let proposal: HomeLayout;
  if (raw.format === HOME_EXPORT_FORMAT) {
    proposal = parseHomeLayout(raw.layout);
  } else if ('dashboardGroups' in raw) {
    const groups = parseLegacyHomeGroups(raw.dashboardGroups);
    proposal = parseHomeLayout(importLegacyHome(current, groups));
  } else {
    throw new Error('Choose a NodeHerder Home or areas export.');
  }
  proposal.etag = current.etag;
  proposal.revision = current.revision;
  return proposal;
}

/** Serialize only placement references/preferences in the explicit Home export format. */
export function serializeHomeExport(layout: HomeLayout): string {
  return JSON.stringify({ format: HOME_EXPORT_FORMAT, layout: cloneHome(layout) }, null, 2);
}
